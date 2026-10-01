<script>
  import {
    DetectDevice,
    FlashFirmware,
    GetUdevRulesContent,
    GetKeyboardImage,
    GetAppIcon,
    SelectFirmware,
    GetAvailableKeyboards,
    GetTranslations,
    GetAvailableLanguages,
    GetVersion
  } from '../wailsjs/go/main/App'

  import {
    EventsOn,
    BrowserOpenURL
  } from '../wailsjs/runtime/runtime'

  import { onMount } from 'svelte'

  let appVersion = $state('')
  let currentLang = $state('en')
  let t = $state({})
  let availableLanguages = $state([])
  let state = $state('idle')
  let device = $state(null)
  let keyboardImage = $state('')
  let appIcon = $state('')
  let logs = $state([])
  let errorMsg = $state('')
  let showLogsModal = $state(false)
  let showKeyboardSelectModal = $state(false)
  let showConfirmModal = $state(false)
  let showAboutModal = $state(false)
  let showUdevWarning = $state(false)
  let isLinux = $state(false)
  let selectedFirmware = $state('')
  let selectedKeyboardModel = $state('')
  let isCustomFirmware = $state(false)
  let confirmationText = $state('')
  let countdown = $state(15)
  let countdownInterval = $state(null)
  let showModelSelectModal = $state(false)
  let modelCandidates = $state([])
  let selectedModelIndex = $state(null)
  let availableKeyboards = $state([])
  let theme = $state('light')
  let iconTapCount = $state(0)
  let iconTapTimer = $state(null)
  let logsCopied = $state(false)
  let copyFeedbackTimer = $state(null)
  let notification = $state('')
  let notificationTimer = $state(null)

  function handleIconTap() {
    iconTapCount++

    if (iconTapTimer) {
      clearTimeout(iconTapTimer)
    }

    if (iconTapCount >= 3) {
      setTheme(theme === 'light' ? 'dark' : 'light')
      iconTapCount = 0
      return
    }

    iconTapTimer = setTimeout(() => {
      iconTapCount = 0
      iconTapTimer = null
    }, 1000)
  }

  function setTheme(newTheme) {
    theme = newTheme
    document.documentElement.setAttribute('data-theme', newTheme)
    localStorage.setItem('theme', newTheme)
  }

  function toggleTheme() {
    setTheme(theme === 'light' ? 'dark' : 'light')
  }

  let canFlash = $derived(
    state === 'ready' &&
    device &&
    device.canFlash !== false &&
    !showUdevWarning &&
    (device.isBootloader ? selectedFirmware : device.firmwarePath)
  )

  let requiresConfirmation = $derived(
    !isCustomFirmware && !device?.isBootloader
  )

  let confirmationMatch = $derived(
    !requiresConfirmation ||
    confirmationText.trim() === selectedKeyboardModel
  )

  let canProceed = $derived(
    countdown === 0 && confirmationMatch
  )

  onMount(async () => {
    logs = []
    isLinux = navigator.platform.toLowerCase().includes('linux')
    appIcon = await GetAppIcon()
    availableKeyboards = await GetAvailableKeyboards()
    availableLanguages = await GetAvailableLanguages()
    t = await GetTranslations(currentLang)
    appVersion = await GetVersion()
    const savedTheme = localStorage.getItem('theme')
    setTheme(savedTheme || 'light')
  })

  $effect(() => {
    const unsubscribe = EventsOn('log', (logEntry) => {
      logs = [...logs, logEntry]
    })

    return unsubscribe
  })

  async function switchLanguage(lang) {
    currentLang = lang
    t = await GetTranslations(lang)
  }

  async function copyUdevRules() {
    const rules = await GetUdevRulesContent()
    await navigator.clipboard.writeText(rules)
  }

  async function copyLogs() {
    try {
      const text = logs.map(log => `[${log.timestamp}] ${log.message}`).join('\n')
      await navigator.clipboard.writeText(text)

      clearTimeout(copyFeedbackTimer)
      logsCopied = true
      copyFeedbackTimer = setTimeout(() => {
        logsCopied = false
        copyFeedbackTimer = null
      }, 1000)
    } catch (err) {
      handleAppError(err)
    }
  }
  function getErrorCode(err) {
    const message = err?.message || ''

    const match = message.match(
      /^(NO_DEVICE|USB_PERMISSION|HID_ENUMERATION|FLASH)(?::|$)/
    )

    return match?.[1] || 'UNKNOWN'
  }

  function notifyError(message) {
    notification = message
    clearTimeout(notificationTimer)
    notificationTimer = setTimeout(() => {
      notification = ''
      notificationTimer = null
    }, 3000)
  }

  function handleAppError(err) {
    const code = getErrorCode(err)

    console.error('Operation failed:', err)

    showUdevWarning =
      code === 'USB_PERMISSION' && isLinux

    switch (code) {
      case 'NO_DEVICE':
        errorMsg = t.NoDeviceDetected
        break

      case 'USB_PERMISSION':
        errorMsg = t.USBPermissionError
        break

      case 'HID_ENUMERATION':
        errorMsg = t.HIDEnumerationError
        break

      case 'FLASH':
        errorMsg = t.FlashError
        break

      default:
        errorMsg = t.UnexpectedError
        break
    }
    notifyError(errorMsg)
  }

  async function detectDevice() {
    state = 'detecting'
    logs = []
    device = null
    keyboardImage = ''
    selectedFirmware = ''
    selectedKeyboardModel = ''
    errorMsg = ''
    showUdevWarning = false
    modelCandidates = []
    selectedModelIndex = null

    try {
      device = await DetectDevice()

      if (device.candidates) {
        modelCandidates = JSON.parse(device.candidates)
        showModelSelectModal = true
        state = 'model-select'
        return
      }

      if (!device.isBootloader && device.firmwarePath) {
        const match = device.firmwarePath.match(/([A-Z0-9]+)\.bin$/i)

        if (match) {
          selectedKeyboardModel = match[1]
        }
      }

      keyboardImage = await GetKeyboardImage(device)
      state = 'ready'

    } catch (err) {
      state = 'error'
      handleAppError(err)
    }
  }

  async function confirmModelSelection() {
    if (selectedModelIndex === null) return

    const model = modelCandidates[selectedModelIndex]

    if (!model.firmwarePath) {
      showModelSelectModal = false
      state = 'error'
      errorMsg = t.NotCurrentlySupported
      return
    }

    device.firmwarePath = model.firmwarePath
    device.name = model.description
    selectedKeyboardModel = model.name
    device.candidates = ''

    showModelSelectModal = false

    keyboardImage = await GetKeyboardImage(device)
    state = 'ready'
  }

  async function selectKeyboard(keyboard) {
    if (!keyboard.firmwarePath) {
      showKeyboardSelectModal = false
      state = 'error'
      errorMsg = t.NotCurrentlySupported
      return
    }

    selectedFirmware = keyboard.firmwarePath
    selectedKeyboardModel = keyboard.name
    isCustomFirmware = false
    showKeyboardSelectModal = false

    keyboardImage = await GetKeyboardImage({
      ...device,
      firmwarePath: keyboard.firmwarePath
    })
  }

  async function browseCustomFirmware() {
    showKeyboardSelectModal = false

    try {
      const path = await SelectFirmware()

      if (path) {
        selectedFirmware = path
        selectedKeyboardModel = path.split('/').pop().replace('.bin', '')
        isCustomFirmware = true
        keyboardImage = ''
      }
    } catch (err) {
      handleAppError(err)
    }
  }

  function startCountdown() {
    countdown = 15

    countdownInterval = setInterval(() => {
      countdown--

      if (countdown <= 0) {
        clearInterval(countdownInterval)
        countdownInterval = null
      }
    }, 1000)
  }

  function openConfirmModal() {
    if (!canFlash) return

    confirmationText = ''
    showConfirmModal = true
    startCountdown()
  }

  function closeConfirmModal() {
    showConfirmModal = false
    confirmationText = ''

    if (countdownInterval) {
      clearInterval(countdownInterval)
      countdownInterval = null
    }
  }

  async function confirmAndFlash() {
    if (!canProceed) return

    closeConfirmModal()
    await flashDevice()
  }

  async function flashDevice() {
    if (!canFlash) return

    state = 'flashing'
    logs = []
    errorMsg = ''
    showUdevWarning = false

    try {
      const fwPath = device.isBootloader ? selectedFirmware : ''

      await FlashFirmware(device, fwPath, 0)

      state = 'success'
    } catch (err) {
      state = 'error'
      handleAppError(err)
    }
  }

  function reset() {
    state = 'idle'
    device = null
    keyboardImage = ''
    selectedFirmware = ''
    selectedKeyboardModel = ''
    logs = []
    errorMsg = ''
    showUdevWarning = false
    isCustomFirmware = false
  }

  function getLogClass(level) {
    switch (level) {
      case 'success':
        return 'text-success'
      case 'error':
        return 'text-error'
      case 'warn':
        return 'text-warning'
      default:
        return 'text-primary-content/80'
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-base-200">
  <div class="w-full max-w-5xl px-6">
    <!-- Main Card -->
    <div class="card bg-base-100 shadow-xl border border-base-300 relative">
      <div class="card-body min-h-[400px] p-6 flex flex-col">
        
        {#if !device && state === 'idle'}
          <!-- Language Toggle - Top Right -->
          <div class="absolute top-4 right-4">
            <div role="tablist" class="tabs tabs-box tabs-sm">
              {#each availableLanguages as lang}
                <button
                  type="button"
                  role="tab"
                  class="tab text-xl px-3 {currentLang === lang.code ? 'tab-active' : ''}"
                  onclick={() => switchLanguage(lang.code)}
                  title={lang.name}
                  aria-label={lang.name}
                  aria-selected={currentLang === lang.code}
                >
                  <span aria-hidden="true" class="emoji">{lang.flag}</span>
                </button>
              {/each}
            </div>
          </div>

          <!-- Initial State -->
          <div class="flex flex-1 items-center justify-center">
            <div class="flex flex-col items-center text-center">
              <button
                type="button"
                class="w-20 h-20 mb-4 bg-base-100 rounded-full flex items-center justify-center"
                onclick={handleIconTap}
              >
                <img
                  src={appIcon}
                  alt={t.AppTitle}
                  class="w-full h-full object-contain rounded-lg"
                />
              </button>
              <p class="text-base-content/70 text-sm mb-6">
                {t.ConnectAndDetect}
              </p>
              <button class="btn btn-neutral btn-wide" onclick={detectDevice}>
                {t.DetectDevice}
              </button>
            </div>
          </div>
        {/if}

        {#if state === 'detecting'}
          <!-- Detecting State -->
          <div class="flex flex-1 items-center justify-center">
            <div class="flex flex-col items-center text-center">
              <span class="loading loading-spinner loading-lg mb-3 text-base-content/70"></span>
              <p class="text-base-content/70 text-sm">{t.Scanning}</p>
            </div>
          </div>
        {/if}

        {#if device && state === 'ready'}
          <!-- Device Detected -->
          <div class="flex-1">
            <div class="grid grid-cols-2 gap-6 mb-4">
              <!-- Left Column: Keyboard Image -->
              <div class="flex items-center justify-center">
                {#if keyboardImage}
                  <img src={keyboardImage} alt={device.name} class="w-full h-auto object-contain" />
                {:else}
                  <div class="w-full h-56 flex items-center justify-center bg-base-200 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 text-base-content/40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                    </svg>
                  </div>
                {/if}
              </div>

              <!-- Right Column: Device Info -->
              <div class="flex flex-col justify-center">
                <h2 class="text-2xl font-light text-base-content mb-3">{device.name}</h2>
                <div class="space-y-1.5 text-sm text-base-content/70">
                  <div class="flex justify-between border-b border-base-300 pb-1.5">
                    <span class="text-base-content/60">VID</span>
                    <span class="font-mono">0x{device.vid}</span>
                  </div>
                  <div class="flex justify-between border-b border-base-300 pb-1.5">
                    <span class="text-base-content/60">PID</span>
                    <span class="font-mono">0x{device.pid}</span>
                  </div>
                  {#if device.manufacturer}
                    <div class="flex justify-between border-b border-base-300 pb-1.5">
                      <span class="text-base-content/60">{t.Manufacturer}</span>
                      <span>{device.manufacturer}</span>
                    </div>
                  {/if}
                  {#if device.product}
                    <div class="flex justify-between border-b border-base-300 pb-1.5">
                      <span class="text-base-content/60">{t.Product}</span>
                      <span>{device.product}</span>
                    </div>
                  {/if}
                  <div class="flex justify-between pt-1.5">
                    <span class="text-base-content/60">{t.Mode}</span>
                    {#if !device.isBootloader}
                      <span class="badge badge-warning badge-sm">{t.ApplicationMode}</span>
                    {:else}
                      <span class="badge badge-success badge-sm">{t.BootloaderMode}</span>
                    {/if}
                  </div>
                  
                  {#if device.isBootloader && selectedFirmware}
                    <div class="flex justify-between border-t border-base-300 pt-1.5 mt-2">
                      <span class="text-base-content/60">{t.Firmware}</span>
                      <span class="text-xs text-success font-mono">{selectedKeyboardModel}</span>
                    </div>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            {#if device.isBootloader}
              <!-- Bootloader Mode: Select Keyboard Button -->
              <button class="btn btn-outline btn-lg w-full mb-2" onclick={() => showKeyboardSelectModal = true}>
                {selectedFirmware ? t.ChangeKeyboard : t.SelectKeyboard}
              </button>
              <button class="btn btn-neutral btn-lg w-full mb-2" onclick={openConfirmModal} disabled={!canFlash}>
                {t.FlashFirmware}
              </button>
            {:else}
              <!-- Application Mode: Direct Flash -->
              <button class="btn btn-neutral btn-lg w-full mb-2" onclick={openConfirmModal} disabled={!canFlash}>
                {t.FlashFirmware}
              </button>
            {/if}

            <!-- Actions -->
            <div class="flex gap-2">
              <button class="btn btn-ghost btn-sm flex-1 text-base-content/70" onclick={detectDevice}>
                {t.DetectAgain}
              </button>
              <button class="btn btn-ghost btn-sm flex-1 text-base-content/70" onclick={() => showLogsModal = true}>
                {t.ShowLogs}
              </button>
            </div>
          </div>
        {/if}

        {#if state === 'flashing'}
          <!-- Flashing State -->
          <div class="flex flex-1 items-center justify-center">
            <div class="flex flex-col items-center text-center">
              <div class="w-20 h-20 mb-4 bg-base-200 rounded-full flex items-center justify-center">
                <span class="loading loading-spinner loading-lg text-base-content/70"></span>
              </div>

              <h3 class="text-lg font-light text-base-content mb-2">
                {t.FlashingFirmware}
              </h3>

              <p class="text-base-content/70 text-sm mb-4">
                {t.DoNotDisconnect}
              </p>

              {#if logs.length > 0}
                <button class="btn btn-ghost btn-wide" onclick={() => showLogsModal = true}>
                  {t.ViewProgress}
                </button>
              {/if}
            </div>
          </div>
        {/if}


        {#if state === 'success'}
          <!-- Success State -->
          <div class="flex flex-1 items-center justify-center">
            <div class="flex flex-col items-center text-center">
              <div class="w-20 h-20 mb-4 bg-success/10 rounded-full flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                </svg>
              </div>

              <h3 class="text-lg font-light text-base-content mb-2">{t.FlashComplete}</h3>
              <p class="text-base-content/70 text-sm mb-4">{t.KeyboardWillReboot}</p>

              <button class="btn btn-outline btn-wide" onclick={reset}>
                {t.FlashAnother}
              </button>
              {#if logs.length > 0}
                <button class="btn btn-ghost btn-wide" onclick={() => showLogsModal = true}>
                  {t.ViewProgress}
                </button>
              {/if}
            </div>
          </div>
        {/if}

        {#if state === 'error'}
          <!-- Error State -->
          <div class="flex flex-1 items-center justify-center">
            <div class="flex flex-col items-center text-center w-full">
              <div class="w-20 h-20 mb-4 bg-error/10 rounded-full flex items-center justify-center">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </div>

              <h3 class="text-lg font-light text-base-content mb-2">
                {errorMsg}
              </h3>

              {#if isLinux && showUdevWarning}
                <div class="alert alert-warning text-left max-w-xl mx-auto mb-4 py-3">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>

                  <div class="flex-1 text-xs">
                    <h3 class="font-semibold">{t.USBPermissionsRequired}</h3>

                    <div class="mt-1">
                      <button class="link link-hover" onclick={copyUdevRules}>
                        {t.CopyRules}
                      </button>

                      {t.ThenPaste}

                      <code class="badge badge-neutral badge-sm">
                        sudo tee /etc/udev/rules.d/50-sonix-keyboards.rules
                      </code>
                    </div>
                  </div>
                </div>
              {/if}

              <p class="text-xs text-error mb-4">{t.CheckLogsForDetails}</p>

              <div class="flex gap-2 justify-center">
                <button class="btn btn-neutral" onclick={detectDevice}>
                  {t.TryAgain}
                </button>

                {#if logs.length > 0}
                  <button class="btn btn-ghost" onclick={() => showLogsModal = true}>
                    {t.ViewLogs}
                  </button>
                {/if}
              </div>
            </div>
          </div>
        {/if}

      </div>
    </div>

    {#if notification}
      <div class="toast toast-top toast-center z-[100]" role="alert" aria-live="assertive">
        <div class="alert alert-error shadow-lg">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{notification}</span>
        </div>
      </div>
    {/if}

    <!-- Footer -->
    <div class="text-center mt-3 text-xs text-base-content/50 pb-4">
      Powered by <button class="link link-hover" onclick={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC')}>SonixFlasher</button>
      <span class="mx-2">·</span>
      <button class="link link-hover" onclick={() => showAboutModal = true}>{t.About}</button>
    </div>

  </div>
</div>

<!-- Keyboard Selection Modal -->
{#if showKeyboardSelectModal}
  <div class="modal modal-open">
    <div class="modal-box bg-base-100">
      <h3 class="font-light text-xl mb-4 text-base-content">{t.SelectKeyboard}</h3>
      
      <div class="space-y-2">
        {#each availableKeyboards as keyboard}
          <button 
            class="btn btn-outline w-full justify-start" 
            onclick={() => selectKeyboard(keyboard)}
          >
            <div class="text-left">
              <div class="font-semibold">{keyboard.name}</div>
              <div class="text-xs opacity-60">{keyboard.description}</div>
            </div>
          </button>
        {/each}
        
        <div class="divider text-xs">or</div>
        
        <button class="btn btn-ghost w-full" onclick={browseCustomFirmware}>
          {t.BrowseCustomFirmware}
        </button>
      </div>

      <div class="modal-action">
        <button class="btn btn-sm" onclick={() => showKeyboardSelectModal = false}>{t.Cancel}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Confirmation Modal -->
{#if showConfirmModal}
  <div class="modal modal-open">
    <div class="modal-box bg-base-100 max-w-md">
      <h3 class="font-bold text-xl text-error mb-4 flex items-center gap-2">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 fill-current" viewBox="0 0 24 24">
          <path d="M12 2.5a2 2 0 0 1 1.732 1l9 15.588A2 2 0 0 1 21 22H3a2 2 0 0 1-1.732-2.912l9-15.588A2 2 0 0 1 12 2.5z M11 9v5h2V9h-2zm0 7v2h2v-2h-2z" fill-rule="evenodd" clip-rule="evenodd"/>
        </svg>
        {t.Warning}
      </h3>
      
      <div class="alert alert-error mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <div class="text-sm">
          <p class="font-bold">{t.DangerousOperation}</p>
          <p class="text-xs mt-1">{t.FlashingWrongFirmware}</p>
        </div>
      </div>

      <div class="mb-4">
        <p class="text-sm text-base-content/70 mb-2">
          {#if isCustomFirmware}
            {t.AboutToFlashCustom}
          {:else}
            {t.AboutToFlashModelPrefix} <strong>{selectedKeyboardModel}</strong>. {t.AboutToFlashModelSuffix}
          {/if}
        </p>
      </div>

      {#if requiresConfirmation}
        <div class="form-control mb-4">
          <label class="label mb-2" for="confirmation-input">
            <span class="label-text text-sm">
              {t.TypePrefix}
              <code class="font-bold bg-base-300 px-1 rounded">{selectedKeyboardModel}</code>
              {t.TypeSuffix}
            </span>
          </label>
          <input
            id="confirmation-input"
            type="text"
            bind:value={confirmationText}
            placeholder={selectedKeyboardModel}
            class="input input-bordered w-full"
          />
        </div>
      {/if}

      <div class="flex items-center justify-between gap-4">
        <div class="text-sm text-base-content/70 min-w-0">
          {#if countdown > 0}
            <span>{t.PleaseWaitPrefix}</span>
            <span class="countdown font-mono text-base-content">
              <span
                style="--value:{countdown}; --digits:2;"
                aria-live="polite"
                aria-label={countdown}
              >
                {countdown}
              </span>
            </span>
            <span>{countdown !== 1 ? t.Seconds : t.Second}{t.PleaseWaitSuffix}</span>
          {:else if requiresConfirmation && !confirmationMatch}
            <span class="text-error">{t.TextDoesntMatch}</span>
          {:else}
            <span class="text-success">{t.ReadyToProceed}</span>
          {/if}
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <button class="btn btn-ghost" onclick={closeConfirmModal}>
            {t.Cancel}
          </button>

          <button
            class="btn btn-error"
            onclick={confirmAndFlash}
            disabled={!canProceed}
          >
            {#if countdown > 0}
              <span class="countdown font-mono text-lg">
                <span
                  style="--value:{countdown}; --digits:2;"
                  aria-live="polite"
                  aria-label={countdown}
                >
                  {countdown}
                </span>
              </span>
            {:else}
              {t.FlashNow}
            {/if}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- About Modal -->
{#if showAboutModal}
  <div class="modal modal-open">
    <div class="modal-box bg-base-100 max-w-md">
      <div class="text-center mb-6">
        <div class="w-20 h-20 mx-auto mb-3">
          {#if appIcon}
            <img src={appIcon} alt={t.AppTitle} class="w-full h-full object-contain rounded-lg" />
          {:else}
            <div class="w-full h-full bg-base-200 rounded-lg flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-base-content/70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
          {/if}
        </div>
        <h2 class="text-2xl font-light text-base-content mb-1">{t.AppTitle}</h2>
        <p class="text-sm text-base-content/60">Version {appVersion}</p>
      </div>

      <div class="space-y-4 text-sm text-base-content/70">
        <div>
          <p class="font-semibold mb-1">{t.About}</p>
          <p class="text-xs">{t.AppDescription}</p>
        </div>

        <div>
          <p class="font-semibold mb-1">{t.Copyright}</p>
          <p class="text-xs">© 2026 DesignedbyGG</p>
          <p class="text-xs">Licensed under GPL-3.0</p>
        </div>

        <div>
          <p class="font-semibold mb-1">{t.BuiltWith}</p>
          <div class="flex gap-2 flex-wrap text-xs">
            <button onclick={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC')} class="badge badge-outline badge-sm link link-hover">SonixFlasher</button>
            <button onclick={() => BrowserOpenURL('https://wails.io')} class="badge badge-outline badge-sm link link-hover">Wails v2</button>
            <button onclick={() => BrowserOpenURL('https://go.dev')} class="badge badge-outline badge-sm link link-hover">Go</button>
            <button onclick={() => BrowserOpenURL('https://svelte.dev')} class="badge badge-outline badge-sm link link-hover">Svelte</button>
            <button onclick={() => BrowserOpenURL('https://daisyui.com')} class="badge badge-outline badge-sm link link-hover">DaisyUI</button>
          </div>
        </div>

        <div>
          <p class="font-semibold mb-1">{t.Links}</p>
          <div class="space-y-1 text-xs">
            <div>
              <button onclick={() => BrowserOpenURL('https://github.com/dexter93/DesignedbyGGUpdater')} class="link link-hover">{t.GitHubRepository}</button>
            </div>
            <div>
              <button onclick={() => BrowserOpenURL('https://github.com/dexter93/DesignedbyGGUpdater/issues')} class="link link-hover">{t.ReportIssue}</button>
            </div>
            <div>
              <button onclick={() => BrowserOpenURL('https://github.com/dexter93/DesignedbyGGUpdater/blob/master/LICENSE')} class="link link-hover">{t.ViewLicense}</button>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-sm" onclick={() => showAboutModal = false}>{t.Close}</button>
      </div>
    </div>
  </div>
{/if}

<!-- Logs Modal -->
{#if showLogsModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-4xl bg-base-100">
      <div class="flex justify-between items-center mb-4">
        <h3 class="font-light text-xl text-base-content">{t.ConsoleOutput}</h3>
        {#if logs.length > 0}
          <button
            class="btn btn-sm btn-ghost"
            onclick={copyLogs}
          >
            {#if logsCopied}
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4 mr-1">
                <path d="M20 6 9 17l-5-5"/>
              </svg>
            {:else}
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4 w-4 mr-1">
                <rect width="14" height="14" x="8" y="8" rx="2" ry="2"/>
                <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>
              </svg>
            {/if}
            {t.Copy}
          </button>
        {/if}
      </div>
      
      {#if logs.length === 0}
        <p class="text-center text-base-content/60 py-8">{t.NoLogsAvailable}</p>
      {:else}
        <div class="mockup-code max-h-96 overflow-y-auto text-sm">
          {#each logs as log}
            <pre class={getLogClass(log.level)}><code><span class="opacity-60">{log.timestamp}</span> {log.message}</code></pre>
          {/each}
        </div>
      {/if}

      <div class="modal-action">
        <button class="btn btn-sm" onclick={() => showLogsModal = false}>{t.Close}</button>
      </div>
    </div>
  </div>
{/if}

{#if showModelSelectModal}
  <div class="modal modal-open">
    <div class="modal-box bg-base-100 max-w-4xl">
      <h3 class="font-light text-xl mb-4 text-base-content">{t.SelectYourKeyboard}</h3>
      
      <p class="text-sm text-base-content/70 mb-6">
        {t.MultipleModelsDetected}
      </p>

      <div class="grid grid-cols-2 gap-4 mb-6">
        {#each modelCandidates as model, index}
          <button 
            class="card border-2 transition-all overflow-hidden {selectedModelIndex === index ? 'border-base-content bg-base-200' : 'border-base-300 hover:border-base-content/50'}"
            onclick={() => selectedModelIndex = index}
          >
            <div class="card-body p-0">
              <div class="w-full h-56 bg-base-200">
                {#if model.firmwarePath}
                  {#await GetKeyboardImage({...device, firmwarePath: model.firmwarePath})}
                    <div class="w-full h-full flex items-center justify-center">
                      <span class="loading loading-spinner loading-md text-base-content/50"></span>
                    </div>
                  {:then imageData}
                    <img src={imageData} alt={model.description} class="w-full h-full object-cover" />
                  {/await}
                {:else}
                  {#await GetKeyboardImage({...device, firmwarePath: '', name: model.description})}
                    <div class="w-full h-full flex items-center justify-center">
                      <span class="loading loading-spinner loading-md text-base-content/50"></span>
                    </div>
                  {:then imageData}
                    {#if imageData}
                      <img src={imageData} alt={model.description} class="w-full h-full object-cover" />
                    {:else}
                      <div class="w-full h-full flex items-center justify-center text-base-content/40 bg-base-200">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-20 w-20" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6v12a2 2 0 002 2z" />
                        </svg>
                      </div>
                    {/if}
                  {/await}
                {/if}
              </div>
              <div class="text-left p-4">
                <div class="font-bold text-base-content mb-1">{model.name}</div>
                <div class="text-xs text-base-content/70 mb-2">{model.description}</div>
                {#if !model.firmwarePath}
                  <div class="badge badge-warning badge-sm">{t.ComingSoon}</div>
                {/if}
              </div>
            </div>
          </button>
        {/each}
      </div>

      <div class="modal-action">
        <button class="btn btn-ghost" onclick={() => { showModelSelectModal = false; state = 'idle'; device = null; selectedModelIndex = null; }}>
          {t.Cancel}
        </button>
        <button 
          class="btn btn-neutral" 
          onclick={confirmModelSelection}
          disabled={selectedModelIndex === null}
        >
          {t.ConfirmSelection}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
  }
 
</style>