<script>
  import { DetectDevice, FlashFirmware, CheckUdevRules, GetUdevRulesContent, GetKeyboardImage, GetAppIcon, SelectFirmware } from '../wailsjs/go/main/App'
  import { EventsOn } from '../wailsjs/runtime/runtime'
  import { BrowserOpenURL } from '../wailsjs/runtime/runtime'
  import { onMount } from 'svelte'

  let state = 'idle'
  let device = null
  let keyboardImage = ''
  let appIcon = ''
  let logs = []
  let errorMsg = ''
  let showLogsModal = false
  let showKeyboardSelectModal = false
  let showConfirmModal = false
  let showAboutModal = false
  let showUdevWarning = false
  let isLinux = false
  let selectedFirmware = ''
  let selectedKeyboardModel = ''
  let isCustomFirmware = false
  let confirmationText = ''
  let countdown = 15
  let countdownInterval = null
  let confirmationRequired = ''

  const availableKeyboards = [
    { name: 'BSK01', description: 'DesignedbyGG Berserker', firmware: 'firmware/BSK01.bin' },
    { name: 'ICL01', description: 'DesignedbyGG Ironclad v1', firmware: 'firmware/ICL01.bin' },
    { name: 'ICL03', description: 'DesignedbyGG Ironclad v3', firmware: 'firmware/ICL03.bin' },
    { name: 'RB01', description: 'DesignedbyGG RedBladeFS', firmware: 'firmware/RB01.bin' },
  ]

  $: canFlash = state === 'ready' && device && !showUdevWarning && (device.isBootloader ? selectedFirmware : device.firmwarePath)
  $: confirmationMatch = confirmationText.trim() === confirmationRequired
  $: canProceed = countdown === 0 && confirmationMatch

  onMount(async () => {
    logs = []
    isLinux = navigator.platform.toLowerCase().includes('linux')
    appIcon = await GetAppIcon()
  })

  async function copyUdevRules() {
    const rules = await GetUdevRulesContent()
    await navigator.clipboard.writeText(rules)
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
    
    try {
      device = await DetectDevice()
      
      // Extract model name FIRST (before loading image)
      if (!device.isBootloader && device.firmwarePath) {
        const match = device.firmwarePath.match(/([A-Z0-9]+)\.bin$/i)
        if (match) {
          selectedKeyboardModel = match[1]
        }
      }
      
      // Then load image
      keyboardImage = await GetKeyboardImage(device)
      state = 'ready'
      
    } catch (err) {
      state = 'error'
      if (err && err.message && (err.message.includes('Permission denied') || err.message.includes('access'))) {
        errorMsg = 'USB Permission Error: Please install udev rules'
        if (isLinux) {
          showUdevWarning = true
        }
      } else {
        errorMsg = err?.message || 'No device detected'
      }
    }
  }

  async function selectKeyboard(keyboard) {
    selectedFirmware = keyboard.firmware
    selectedKeyboardModel = keyboard.name
    isCustomFirmware = false
    showKeyboardSelectModal = false
    
    // Update image by passing device with the selected firmware path
    keyboardImage = await GetKeyboardImage({
      ...device,
      firmwarePath: keyboard.firmware
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
      console.error('Failed to select firmware:', err)
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
    
    confirmationRequired = isCustomFirmware ? "I know what I'm doing, I'm on my own" : selectedKeyboardModel
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
      const result = await FlashFirmware(device, fwPath, 0)
      
      if (result && result.message === 'USB_PERMISSION_ERROR') {
        state = 'ready'
        errorMsg = 'USB Permission Error - Udev rules required'
        if (isLinux) {
          showUdevWarning = true
        }
        return
      }
      
      if (result && result.success === true) {
        state = 'success'
      } else {
        state = 'error'
        errorMsg = result?.message || 'Flash operation failed'
      }
    } catch (err) {
      state = 'error'
      errorMsg = err?.message || 'Flash failed - check logs for details'
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
    switch(level) {
      case 'success': return 'text-success'
      case 'error': return 'text-error'
      case 'warn': return 'text-warning'
      default: return 'text-neutral/70'
    }
  }

  EventsOn('log', (logEntry) => {
    logs = [...logs, logEntry]
  })
</script>

<div class="h-screen flex items-center justify-center bg-neutral-50" data-theme="light">
  <div class="w-full max-w-5xl px-6">
    <!-- Main Card -->
    <div class="card bg-white shadow-xl border border-neutral-200">
      <div class="card-body p-6">
        
        {#if !device && state === 'idle'}
          <!-- Initial State -->
          <div class="text-center py-16">
            <div class="w-20 h-20 mx-auto mb-4 bg-neutral-100 rounded-full flex items-center justify-center">
              <img src={appIcon} alt="DesignedbyGG Updater" class="w-full h-full object-contain rounded-lg" />
            </div>
            <p class="text-neutral-600 text-sm mb-6">Connect your keyboard and click detect</p>
            <button class="btn btn-neutral btn-wide" on:click={detectDevice}>
              Detect Device
            </button>
          </div>
        {/if}

        {#if state === 'detecting'}
          <!-- Detecting State -->
          <div class="text-center py-16">
            <span class="loading loading-spinner loading-lg mb-3 text-neutral-600"></span>
            <p class="text-neutral-600 text-sm">Scanning for devices...</p>
          </div>
        {/if}

        {#if device && state === 'ready'}
          <!-- Device Detected -->
          <div>
            <div class="grid grid-cols-2 gap-6 mb-4">
              <!-- Left Column: Keyboard Image -->
              <div class="flex items-center justify-center">
                {#if keyboardImage}
                  <img src={keyboardImage} alt={device.name} class="w-full h-auto object-contain" />
                {:else}
                  <div class="w-full h-56 flex items-center justify-center bg-neutral-50 rounded-lg">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 text-neutral-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                    </svg>
                  </div>
                {/if}
              </div>

              <!-- Right Column: Device Info -->
              <div class="flex flex-col justify-center">
                <h2 class="text-2xl font-light text-neutral-800 mb-3">{device.name}</h2>
                <div class="space-y-1.5 text-sm text-neutral-600">
                  <div class="flex justify-between border-b border-neutral-100 pb-1.5">
                    <span class="text-neutral-500">VID</span>
                    <span class="font-mono">0x{device.vid}</span>
                  </div>
                  <div class="flex justify-between border-b border-neutral-100 pb-1.5">
                    <span class="text-neutral-500">PID</span>
                    <span class="font-mono">0x{device.pid}</span>
                  </div>
                  {#if device.manufacturer}
                    <div class="flex justify-between border-b border-neutral-100 pb-1.5">
                      <span class="text-neutral-500">Manufacturer</span>
                      <span>{device.manufacturer}</span>
                    </div>
                  {/if}
                  {#if device.product}
                    <div class="flex justify-between border-b border-neutral-100 pb-1.5">
                      <span class="text-neutral-500">Product</span>
                      <span>{device.product}</span>
                    </div>
                  {/if}
                  <div class="flex justify-between pt-1.5">
                    <span class="text-neutral-500">Mode</span>
                    {#if !device.isBootloader}
                      <span class="badge badge-warning badge-sm">Application</span>
                    {:else}
                      <span class="badge badge-success badge-sm">Bootloader</span>
                    {/if}
                  </div>
                  
                  {#if device.isBootloader && selectedFirmware}
                    <div class="flex justify-between border-t border-neutral-100 pt-1.5 mt-2">
                      <span class="text-neutral-500">Firmware</span>
                      <span class="text-xs text-success font-mono">{selectedKeyboardModel}</span>
                    </div>
                  {/if}
                </div>
              </div>
            </div>

            <!-- Action Buttons -->
            {#if isLinux && showUdevWarning}
              <!-- Udev Warning Block -->
              <div class="alert alert-warning mb-2 py-3">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-5 h-5"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                <div class="flex-1 text-xs">
                  <h3 class="font-semibold">USB Permissions Required</h3>
                  <div class="mt-1">
                    <button class="link link-hover" on:click={copyUdevRules}>Copy rules</button>
                    then: <code class="text-xs bg-neutral-800 text-neutral-200 px-1 py-0.5 rounded">sudo tee /etc/udev/rules.d/50-sonix-keyboards.rules</code>
                  </div>
                </div>
              </div>
              <button class="btn btn-ghost btn-xs w-full mb-2" on:click={() => showUdevWarning = false}>
                Continue anyway
              </button>
            {:else if device.isBootloader}
              <!-- Bootloader Mode: Select Keyboard Button -->
              <button class="btn btn-outline btn-lg w-full mb-2" on:click={() => showKeyboardSelectModal = true}>
                {selectedFirmware ? 'Change Keyboard' : 'Select Keyboard'}
              </button>
              <button class="btn btn-neutral btn-lg w-full mb-2" on:click={openConfirmModal} disabled={!canFlash}>
                Flash Firmware
              </button>
            {:else}
              <!-- Application Mode: Direct Flash -->
              <button class="btn btn-neutral btn-lg w-full mb-2" on:click={openConfirmModal}>
                Flash Firmware
              </button>
            {/if}

            <!-- Actions -->
            <div class="flex gap-2">
              <button class="btn btn-ghost btn-sm flex-1 text-neutral-600" on:click={detectDevice}>
                Detect Again
              </button>
              <button class="btn btn-ghost btn-sm flex-1 text-neutral-600" on:click={() => showLogsModal = true}>
                Show Logs
              </button>
            </div>
          </div>
        {/if}

        {#if state === 'flashing'}
          <!-- Flashing State -->
          <div class="text-center py-16">
            <div class="mb-3 flex justify-center">
              <span class="loading loading-spinner loading-lg text-neutral-600"></span>
            </div>
            <p class="text-neutral-600">Flashing firmware...</p>
            <p class="text-xs text-neutral-400 mt-1">Do not disconnect</p>
            {#if logs.length > 0}
              <button class="btn btn-ghost btn-sm mt-4" on:click={() => showLogsModal = true}>
                View Progress
              </button>
            {/if}
          </div>
        {/if}

        {#if state === 'success'}
          <!-- Success State -->
          <div class="text-center py-16">
            <div class="w-20 h-20 mx-auto mb-4 bg-green-50 rounded-full flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h3 class="text-lg font-light text-neutral-800 mb-2">Flash Complete</h3>
            <p class="text-neutral-600 text-sm mb-4">Your keyboard will reboot automatically</p>
            <button class="btn btn-outline btn-wide" on:click={reset}>
              Flash Another
            </button>
          </div>
        {/if}

        {#if state === 'error'}
          <!-- Error State -->
          <div class="text-center py-16">
            <div class="w-20 h-20 mx-auto mb-4 bg-red-50 rounded-full flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h3 class="text-lg font-light text-neutral-800 mb-2">
              {errorMsg.includes('No device') || errorMsg.includes('not detected') ? 'No Device Found' : 'Flash Failed'}
            </h3>
            <p class="text-xs text-red-600 mb-4">{errorMsg}</p>
            <div class="flex gap-2 justify-center">
              <button class="btn btn-neutral" on:click={detectDevice}>
                Try Again
              </button>
              {#if logs.length > 0}
                <button class="btn btn-ghost" on:click={() => showLogsModal = true}>
                  View Logs
                </button>
              {/if}
            </div>
          </div>
        {/if}

      </div>
    </div>

    <!-- Footer -->
    <div class="text-center mt-3 text-xs text-neutral-400 pb-4">
      Powered by <button class="link link-hover" on:click={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC')}>SonixFlasher</button>
      <span class="mx-2">·</span>
      <button class="link link-hover" on:click={() => showAboutModal = true}>About</button>
    </div>

  </div>
</div>

<!-- Keyboard Selection Modal -->
{#if showKeyboardSelectModal}
  <div class="modal modal-open">
    <div class="modal-box bg-white">
      <h3 class="font-light text-xl mb-4 text-neutral-800">Select Keyboard</h3>
      
      <div class="space-y-2">
        {#each availableKeyboards as keyboard}
          <button 
            class="btn btn-outline w-full justify-start" 
            on:click={() => selectKeyboard(keyboard)}
          >
            <div class="text-left">
              <div class="font-semibold">{keyboard.name}</div>
              <div class="text-xs opacity-60">{keyboard.description}</div>
            </div>
          </button>
        {/each}
        
        <div class="divider text-xs">or</div>
        
        <button class="btn btn-ghost w-full" on:click={browseCustomFirmware}>
          Browse for custom firmware...
        </button>
      </div>

      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => showKeyboardSelectModal = false}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<!-- Confirmation Modal -->
{#if showConfirmModal}
  <div class="modal modal-open">
    <div class="modal-box bg-white max-w-md">
      <h3 class="font-bold text-xl text-error mb-4">⚠️ Warning</h3>
      
      <div class="alert alert-error mb-4">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="stroke-current shrink-0 w-6 h-6"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
        <div class="text-sm">
          <p class="font-bold">This is a potentially dangerous operation!</p>
          <p class="text-xs mt-1">Flashing wrong firmware can brick your device.</p>
        </div>
      </div>

      <div class="mb-4">
        <p class="text-sm text-neutral-600 mb-2">
          {#if isCustomFirmware}
            You are about to flash <strong>custom firmware</strong>. You are responsible for any issues that may occur.
          {:else}
            You are about to flash firmware for <strong>{selectedKeyboardModel}</strong>. Make sure this matches your keyboard model.
          {/if}
        </p>
      </div>

      <div class="form-control mb-4">
        <label class="label" for="confirmation-input">
          <span class="label-text text-sm">Type <code class="font-bold bg-neutral-200 px-1 rounded">{confirmationRequired}</code> to proceed:</span>
        </label>
        <input 
          id="confirmation-input"
          type="text" 
          bind:value={confirmationText}
          placeholder={confirmationRequired}
          class="input input-bordered w-full"
          disabled={countdown > 0}
        />
      </div>

      <div class="flex items-center justify-between mb-4">
        <div class="text-sm text-neutral-600">
          {#if countdown > 0}
            Please wait {countdown} second{countdown !== 1 ? 's' : ''}...
          {:else if !confirmationMatch}
            <span class="text-error">Text doesn't match</span>
          {:else}
            <span class="text-success">Ready to proceed</span>
          {/if}
        </div>
        {#if countdown > 0}
          <div class="radial-progress text-warning" style="--value:{((15 - countdown) / 15 * 100).toFixed(0)}; --size:2.5rem; --thickness: 3px;">
            {countdown}
          </div>
        {/if}
      </div>

      <div class="modal-action">
        <button class="btn btn-ghost" on:click={closeConfirmModal}>Cancel</button>
        <button class="btn btn-error" on:click={confirmAndFlash} disabled={!canProceed}>
          Flash Now
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- About Modal -->
{#if showAboutModal}
  <div class="modal modal-open">
    <div class="modal-box bg-white max-w-md">
      <div class="text-center mb-6">
        <div class="w-20 h-20 mx-auto mb-3">
          {#if appIcon}
            <img src={appIcon} alt="DesignedbyGG Updater" class="w-full h-full object-contain rounded-lg" />
          {:else}
            <div class="w-full h-full bg-neutral-100 rounded-lg flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-neutral-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
          {/if}
        </div>
        <h2 class="text-2xl font-light text-neutral-800 mb-1">DesignedbyGG Updater</h2>
        <p class="text-sm text-neutral-500">Version 1.0.0</p>
      </div>

      <div class="space-y-4 text-sm text-neutral-600">
        <div>
          <p class="font-semibold mb-1">About</p>
          <p class="text-xs">A modern GUI wrapper for SonixFlasher, designed to simplify firmware updates for DesignedbyGG keyboards.</p>
        </div>

        <div>
          <p class="font-semibold mb-1">Copyright</p>
          <p class="text-xs">© 2026 DesignedbyGG</p>
          <p class="text-xs">Licensed under GPL-3.0</p>
        </div>

        <div>
          <p class="font-semibold mb-1">Built With</p>
          <div class="flex gap-2 flex-wrap text-xs">
            <button on:click={() => BrowserOpenURL('https://wails.io')} class="badge badge-outline badge-sm link link-hover">Wails v2</button>
            <button on:click={() => BrowserOpenURL('https://go.dev')} class="badge badge-outline badge-sm link link-hover">Go</button>
            <button on:click={() => BrowserOpenURL('https://svelte.dev')} class="badge badge-outline badge-sm link link-hover">Svelte</button>
            <button on:click={() => BrowserOpenURL('https://daisyui.com')} class="badge badge-outline badge-sm link link-hover">DaisyUI</button>
          </div>
        </div>

          <div>
            <p class="font-semibold mb-1">Links</p>
            <div class="space-y-1 text-xs">
              <div>
                <button on:click={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC')} class="link link-hover">GitHub Repository</button>
              </div>
              <div>
                <button on:click={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC/issues')} class="link link-hover">Report an Issue</button>
              </div>
              <div>
                <button on:click={() => BrowserOpenURL('https://github.com/SonixQMK/SonixFlasherC/blob/main/LICENSE')} class="link link-hover">View License</button>
              </div>
            </div>
          </div>
        </div>

      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => showAboutModal = false}>Close</button>
      </div>
    </div>
  </div>
{/if}

<!-- Logs Modal -->
{#if showLogsModal}
  <div class="modal modal-open">
    <div class="modal-box max-w-4xl bg-white">
      <div class="flex justify-between items-center mb-4">
        <h3 class="font-light text-xl text-neutral-800">Console Output</h3>
        {#if logs.length > 0}
          <button class="btn btn-sm btn-ghost" on:click={() => {
            const text = logs.map(l => `[${l.timestamp}] ${l.message}`).join('\n')
            navigator.clipboard.writeText(text)
          }}>
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            Copy
          </button>
        {/if}
      </div>
      
      {#if logs.length === 0}
        <p class="text-center text-neutral-500 py-8">No logs available</p>
      {:else}
        <div class="bg-neutral-50 rounded-lg p-4 max-h-96 overflow-y-auto border border-neutral-200">
          <div class="font-mono text-xs space-y-0.5">
            {#each logs as log}
              <div class={getLogClass(log.level)}>
                <span class="opacity-50">{log.timestamp}</span>
                <span class="ml-2">{log.message}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => showLogsModal = false}>Close</button>
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
  
  :global(html) {
    background: #fafafa;
  }
</style>