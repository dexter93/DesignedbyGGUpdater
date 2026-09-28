package main

import (
	"fmt"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type FlashResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (a *App) GetSonixFlasherPath() (string, error) {
	a.emitLog("info", fmt.Sprintf("Detecting platform: %s/%s", runtime.GOOS, runtime.GOARCH))

	var binaryPath string
	switch runtime.GOOS {
	case "darwin":
		binaryPath = "binaries/darwin/sonixflasher"
		a.emitLog("info", "Using macOS binary")
	case "linux":
		binaryPath = "binaries/linux/sonixflasher"
		a.emitLog("info", "Using Linux binary")
	case "windows":
		binaryPath = "binaries/windows/sonixflasher.exe"
		a.emitLog("info", "Using Windows binary")
	default:
		err := fmt.Errorf("unsupported platform: %s", runtime.GOOS)
		a.emitLog("error", err.Error())
		return "", err
	}

	tmpDir := os.TempDir()

	execName := "sonixflasher"
	if runtime.GOOS == "windows" {
		execName = "sonixflasher.exe"
	}
	tmpPath, err := a.extractEmbeddedFile(binaryPath, tmpDir, execName, 0755)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		if _, err := a.extractEmbeddedFile("binaries/windows/libusb-1.0.dll", tmpDir, "libusb-1.0.dll", 0755); err != nil {
			return "", err
		}
	}

	a.emitLog("success", "Binary extracted successfully")
	return tmpPath, nil
}

func (a *App) extractEmbeddedFile(srcPath, dstDir, dstName string, perm os.FileMode) (string, error) {
	a.emitLog("info", fmt.Sprintf("Extracting embedded binary: %s", srcPath))
	data, err := binaries.ReadFile(srcPath)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to read embedded binary: %v", err))
		return "", fmt.Errorf("failed to read embedded binary: %v", err)
	}

	dstPath := filepath.Join(dstDir, dstName)
	a.emitLog("info", fmt.Sprintf("Writing binary to: %s", dstPath))
	if err := os.WriteFile(dstPath, data, perm); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to write binary: %v", err))
		return "", fmt.Errorf("failed to write binary: %v", err)
	}

	return dstPath, nil
}

func (a *App) GetEmbeddedFirmware(firmwarePath string) (string, error) {
	a.emitLog("info", fmt.Sprintf("Loading embedded firmware: %s", firmwarePath))

	data, err := binaries.ReadFile(firmwarePath)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Firmware not found: %v", err))
		return "", fmt.Errorf("firmware not found: %v", err)
	}

	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, filepath.Base(firmwarePath))

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to write firmware: %v", err))
		return "", fmt.Errorf("failed to write firmware: %v", err)
	}

	a.emitLog("success", fmt.Sprintf("Firmware extracted: %.2f KB", float64(len(data))/1024))
	return tmpPath, nil
}

func (a *App) SelectFirmware() (string, error) {
	a.emitLog("info", "Opening firmware file picker...")

	file, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Firmware",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Binary Files (*.bin)", Pattern: "*.bin"},
		},
	})

	if err != nil {
		a.emitLog("error", fmt.Sprintf("File picker error: %v", err))
		return "", err
	}

	if file == "" {
		a.emitLog("warn", "File selection cancelled")
		return "", nil
	}

	info, err := os.Stat(file)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Cannot access file: %v", err))
		return "", err
	}

	a.emitLog("success", "═══════════════════════════════════════")
	a.emitLog("success", "✓ FIRMWARE SELECTED")
	a.emitLog("info", fmt.Sprintf("  File: %s", filepath.Base(file)))
	a.emitLog("info", fmt.Sprintf("  Path: %s", file))
	a.emitLog("info", fmt.Sprintf("  Size: %d bytes (%.2f KB)", info.Size(), float64(info.Size())/1024))
	a.emitLog("success", "═══════════════════════════════════════")

	return file, nil
}

func (a *App) FlashFirmware(device *Device, customFirmwarePath string, offset int64) (*FlashResult, error) {
	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", "STARTING FLASH OPERATION")
	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", fmt.Sprintf("Device: %s", device.Name))
	a.emitLog("info", fmt.Sprintf("VID:PID: %s/%s", device.VID, device.PID))

	var firmwarePath string
	var cleanupFirmware bool

	if customFirmwarePath != "" {
		if strings.HasPrefix(customFirmwarePath, "firmware/") {
			tmpPath, err := a.GetEmbeddedFirmware(customFirmwarePath)
			if err != nil {
				return nil, newAppError(errCodeFlash, err)
			}
			firmwarePath = tmpPath
			cleanupFirmware = true
			a.emitLog("info", fmt.Sprintf("Using embedded firmware: %s", filepath.Base(customFirmwarePath)))
		} else {
			firmwarePath = customFirmwarePath
			cleanupFirmware = false
			a.emitLog("info", fmt.Sprintf("Using custom firmware: %s", filepath.Base(firmwarePath)))
		}
	} else if device.FirmwarePath != "" {
		tmpPath, err := a.GetEmbeddedFirmware(device.FirmwarePath)
		if err != nil {
			return nil, newAppError(errCodeFlash, err)
		}
		firmwarePath = tmpPath
		cleanupFirmware = true
		a.emitLog("info", fmt.Sprintf("Using embedded firmware: %s", device.FirmwarePath))
	} else {
		err := fmt.Errorf("no firmware specified")
		a.emitLog("error", err.Error())
		return nil, newAppError(errCodeFlash, err)
	}

	if cleanupFirmware {
		defer os.Remove(firmwarePath)
	}

	a.emitLog("info", fmt.Sprintf("Offset: 0x%x (%d bytes)", offset, offset))
	a.emitLog("info", "═══════════════════════════════════════")

	binPath, err := a.GetSonixFlasherPath()
	if err != nil {
		return nil, newAppError(errCodeFlash, err)
	}
	defer os.Remove(binPath)

	vidpid := fmt.Sprintf("%s/%s", device.VID, device.PID)
	args := []string{
		"--vid-pid", vidpid,
		"--file", firmwarePath,
	}

	if !device.IsBootloader {
		args = append(args, "--reboot", "sonix")
		a.emitLog("warn", "Device in application mode, will reboot to bootloader...")
	}

	if offset > 0 {
		args = append(args, "--offset", fmt.Sprintf("0x%x", offset))
	}

	a.emitLog("info", fmt.Sprintf("Executing: %s %s", filepath.Base(binPath), strings.Join(args, " ")))

	cmd := exec.Command(binPath, args...)
	hideConsoleWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to create stdout pipe: %v", err))
		return nil, newAppError(errCodeFlash, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to create stderr pipe: %v", err))
		return nil, newAppError(errCodeFlash, err)
	}

	a.emitLog("info", "Starting sonixflasher process...")
	if err := cmd.Start(); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to start process: %v", err))
		return nil, newAppError(errCodeFlash, err)
	}

	var outputMu sync.Mutex
	successDetected := false
	permissionError := false
	deviceOpenError := false

	handleOutput := func(line string) {
		lower := strings.ToLower(line)

		outputMu.Lock()
		switch {
		case line == "=== FLASHING COMPLETED SUCCESSFULLY ===":
			successDetected = true

		// v3: device was found, but its USB/HID interface could not be opened.
		// This is the flasher's explicit permissions/access path.
		case strings.Contains(lower, "device present but failed to open (permissions?)"),
			strings.Contains(lower, "permission denied"),
			strings.Contains(lower, "access denied"),
			strings.Contains(lower, "access is denied"),
			strings.Contains(lower, "libusb_error_access"):
			permissionError = true

		// v3: device was not found, or all attempts to open it failed without
		// identifying the permissions/access path above.
		case strings.Contains(lower, "device not found"),
			strings.Contains(lower, "failed to open device after"):
			deviceOpenError = true
		}
		outputMu.Unlock()

		level := "info"
		switch {
		case strings.Contains(lower, "failed") ||
			strings.Contains(lower, "error") ||
			strings.Contains(lower, "invalid") ||
			strings.Contains(lower, "mismatch") ||
			strings.Contains(lower, "cannot open") ||
			strings.Contains(lower, "unsupported"):
			level = "error"
		case strings.Contains(lower, "warning") ||
			strings.Contains(lower, "potentially dangerous") ||
			strings.Contains(lower, "skipped"):
			level = "warn"
		case strings.Contains(lower, "successfully") ||
			strings.Contains(lower, "verified"):
			level = "success"
		}

		a.emitLog(level, line)
	}

	outputDone := make(chan bool, 2)

	go func() {
		defer func() { outputDone <- true }()
		_ = scanOutput(stdout, handleOutput)
	}()

	go func() {
		defer func() { outputDone <- true }()
		_ = scanOutput(stderr, handleOutput)
	}()

	<-outputDone
	<-outputDone

	a.emitLog("info", "Waiting for flash process to complete...")
	err = cmd.Wait()

	outputMu.Lock()
	completedSuccessfully := successDetected
	hasPermissionError := permissionError
	hasDeviceOpenError := deviceOpenError
	outputMu.Unlock()

	a.emitLog("info", "═══════════════════════════════════════")

	if hasPermissionError {
		a.emitLog("error", "✗ USB PERMISSION ERROR")
		a.emitLog("error", "═══════════════════════════════════════")

		if runtime.GOOS == "linux" {
			a.emitLog("error", "Udev rules are required for USB access")
			a.emitLog("warn", "Install udev rules and reload:")
			a.emitLog("warn", "  sudo tee /etc/udev/rules.d/50-sonix-keyboards.rules")
			a.emitLog("warn", "  sudo udevadm control --reload-rules")
			a.emitLog("warn", "  sudo udevadm trigger")
		}

		return nil, newAppError(errCodeUSBPermission, nil)
	}

	if hasDeviceOpenError {
		a.emitLog("error", "✗ DEVICE OPEN FAILED")
		a.emitLog("error", "═══════════════════════════════════════")

		err := fmt.Errorf("device could not be opened")
		return nil, newAppError(errCodeFlash, err)
	}

	// SonixFlasherC v3 returns 0 after this exact final marker.
	if err == nil && completedSuccessfully {
		a.emitLog("success", "✓ FLASH COMPLETED SUCCESSFULLY")
		a.emitLog("success", "═══════════════════════════════════════")
		a.emitLog("info", "Device will reboot automatically")
		a.emitLog("info", "You can now disconnect your keyboard")

		return &FlashResult{
			Success: true,
			Message: "Device successfully flashed!",
		}, nil
	}

	if err != nil {
		a.emitLog("error", "✗ FLASH FAILED")
		a.emitLog("error", fmt.Sprintf("Exit error: %v", err))
		a.emitLog("error", "═══════════════════════════════════════")

		return nil, newAppError(errCodeFlash, err)
	}

	a.emitLog("warn", "⚠ FLASH STATUS UNKNOWN")
	a.emitLog("warn", "Process completed but success not confirmed")
	a.emitLog("warn", "═══════════════════════════════════════")

	return nil, newAppError(
		errCodeFlash,
		fmt.Errorf("flash status could not be confirmed"),
	)
}
