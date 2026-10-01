package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type FlashResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type outputStatus struct {
	successDetected bool
	permissionError bool
	deviceOpenError bool
}

func sonixFlasherEmbeddedPath() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "binaries/darwin/sonixflasher", nil
	case "linux":
		return "binaries/linux/sonixflasher", nil
	case "windows":
		return "binaries/windows/sonixflasher.exe", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func sonixFlasherFileName() string {
	if runtime.GOOS == "windows" {
		return "sonixflasher.exe"
	}

	return "sonixflasher"
}

func writeEmbeddedFile(srcPath, dstPath string, perm os.FileMode) error {
	data, err := binaries.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read embedded file %q: %w", srcPath, err)
	}

	if err := os.WriteFile(dstPath, data, perm); err != nil {
		return fmt.Errorf("write embedded file %q: %w", dstPath, err)
	}

	return nil
}

func extractSonixFlasher(dstDir string) (string, error) {
	if strings.TrimSpace(dstDir) == "" {
		return "", fmt.Errorf("temporary directory is required")
	}

	srcPath, err := sonixFlasherEmbeddedPath()
	if err != nil {
		return "", err
	}

	binaryPath := filepath.Join(dstDir, sonixFlasherFileName())
	if err := writeEmbeddedFile(srcPath, binaryPath, 0o755); err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		dllPath := filepath.Join(dstDir, "libusb-1.0.dll")
		if err := writeEmbeddedFile("binaries/windows/libusb-1.0.dll", dllPath, 0o644); err != nil {
			return "", err
		}
	}

	return binaryPath, nil
}

func knownEmbeddedFirmware(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}

	for _, devices := range knownAppModePIDs {
		for _, device := range devices {
			if device.FirmwarePath == path {
				return true
			}
		}
	}

	return false
}

func extractEmbeddedFirmware(dstDir, firmwarePath string) (string, error) {
	if strings.TrimSpace(dstDir) == "" {
		return "", fmt.Errorf("temporary directory is required")
	}
	if !knownEmbeddedFirmware(firmwarePath) {
		return "", fmt.Errorf("unknown embedded firmware %q", firmwarePath)
	}

	dstPath := filepath.Join(dstDir, filepath.Base(firmwarePath))
	if err := writeEmbeddedFile(firmwarePath, dstPath, 0o644); err != nil {
		return "", err
	}

	return dstPath, nil
}

func (a *App) SelectFirmware() (string, error) {
	if a == nil {
		return "", newAppError(errCodeUnavailable, fmt.Errorf("application is unavailable"))
	}

	a.emitLog("info", "Opening firmware file picker...")

	file, err := wailsruntime.OpenFileDialog(a.applicationContext(), wailsruntime.OpenDialogOptions{
		Title: "Select Firmware",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: "Binary Files (*.bin)", Pattern: "*.bin"},
		},
	})
	if err != nil {
		a.emitLog("error", fmt.Sprintf("File picker error: %v", err))
		return "", newAppError(errCodeFlash, err)
	}

	if file == "" {
		a.emitLog("warn", "File selection cancelled")
		return "", nil
	}

	if err := validateCustomFirmware(file); err != nil {
		a.emitLog("error", fmt.Sprintf("Cannot use selected firmware: %v", err))
		return "", newAppError(errCodeInvalidInput, err)
	}

	info, err := os.Stat(file)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Cannot access selected firmware: %v", err))
		return "", newAppError(errCodeFlash, err)
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
	if err := a.requireHID(); err != nil {
		return nil, err
	}
	if err := validateDevice(device); err != nil {
		return nil, newAppError(errCodeInvalidInput, err)
	}
	if offset < 0 {
		return nil, newAppError(errCodeInvalidInput, fmt.Errorf("offset cannot be negative"))
	}
	if err := a.beginFlash(); err != nil {
		return nil, err
	}
	defer a.endFlash()

	resolvedDevice, err := resolveConnectedDevice(device)
	if err != nil {
		var appErr *appError
		if errors.As(err, &appErr) {
			return nil, appErr
		}
		return nil, newAppError(errCodeNoDevice, err)
	}
	device = resolvedDevice

	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", "STARTING FLASH OPERATION")
	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", fmt.Sprintf("Device: %s", device.Name))
	a.emitLog("info", fmt.Sprintf("VID:PID: %s/%s", device.VID, device.PID))

	jobDir, err := os.MkdirTemp("", appName)
	if err != nil {
		return nil, newAppError(errCodeFlash, fmt.Errorf("create temporary directory: %w", err))
	}
	defer func() {
		if err := os.RemoveAll(jobDir); err != nil {
			a.emitLog("warn", fmt.Sprintf("Failed to remove temporary files: %v", err))
		}
	}()

	firmwarePath, err := a.resolveFirmware(jobDir, device, customFirmwarePath)
	if err != nil {
		return nil, err
	}

	binPath, err := extractSonixFlasher(jobDir)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to prepare sonixflasher: %v", err))
		return nil, newAppError(errCodeFlash, err)
	}

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

	a.emitLog("info", fmt.Sprintf("Offset: 0x%x (%d bytes)", offset, offset))
	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", fmt.Sprintf("Executing: %s %s", filepath.Base(binPath), strings.Join(args, " ")))

	ctx, cancel := context.WithTimeout(a.applicationContext(), flashOperationTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, args...)
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

	status := &outputStatus{}
	var statusMu sync.Mutex
	handleOutput := func(line string) {
		updateOutputStatus(status, &statusMu, line)
		a.emitLog(outputLogLevel(line), line)
	}

	scanResults := make(chan error, 2)
	go func() { scanResults <- scanOutput(stdout, handleOutput) }()
	go func() { scanResults <- scanOutput(stderr, handleOutput) }()

	a.emitLog("info", "Waiting for flash process to complete...")
	var scanErrs [2]error
	for i := range scanErrs {
		scanErrs[i] = <-scanResults
		if scanErrs[i] != nil {
			// A scanner that stops draining a pipe can deadlock the child process.
			// Cancel it before waiting for the remaining stream or process exit.
			cancel()
		}
	}

	waitErr := cmd.Wait()

	if ctx.Err() == context.DeadlineExceeded {
		a.emitLog("error", "✗ FLASH TIMED OUT")
		return nil, newAppError(errCodeTimeout, fmt.Errorf("flash operation exceeded %s", flashOperationTimeout))
	}

	if outputErr := errors.Join(scanErrs[0], scanErrs[1]); outputErr != nil {
		a.emitLog("error", fmt.Sprintf("Failed to read sonixflasher output: %v", outputErr))
		return nil, newAppError(errCodeFlash, outputErr)
	}

	statusMu.Lock()
	resultStatus := *status
	statusMu.Unlock()

	return a.flashResult(waitErr, resultStatus)
}

func (a *App) resolveFirmware(jobDir string, device *Device, customFirmwarePath string) (string, error) {
	if customFirmwarePath != "" {
		if knownEmbeddedFirmware(customFirmwarePath) {
			firmwarePath, err := extractEmbeddedFirmware(jobDir, customFirmwarePath)
			if err != nil {
				return "", newAppError(errCodeFlash, err)
			}
			a.emitLog("info", fmt.Sprintf("Using embedded firmware: %s", filepath.Base(customFirmwarePath)))
			return firmwarePath, nil
		}

		if err := validateCustomFirmware(customFirmwarePath); err != nil {
			return "", newAppError(errCodeInvalidInput, err)
		}
		a.emitLog("info", fmt.Sprintf("Using custom firmware: %s", filepath.Base(customFirmwarePath)))
		return customFirmwarePath, nil
	}

	if !knownEmbeddedFirmware(device.FirmwarePath) {
		return "", newAppError(errCodeInvalidInput, fmt.Errorf("no supported embedded firmware specified"))
	}

	firmwarePath, err := extractEmbeddedFirmware(jobDir, device.FirmwarePath)
	if err != nil {
		return "", newAppError(errCodeFlash, err)
	}

	a.emitLog("info", fmt.Sprintf("Using embedded firmware: %s", device.FirmwarePath))
	return firmwarePath, nil
}

func updateOutputStatus(status *outputStatus, mu *sync.Mutex, line string) {
	if status == nil || mu == nil {
		return
	}

	lower := strings.ToLower(line)

	mu.Lock()
	defer mu.Unlock()

	switch {
	case line == "=== FLASHING COMPLETED SUCCESSFULLY ===":
		status.successDetected = true
	case strings.Contains(lower, "device present but failed to open (permissions?)"),
		strings.Contains(lower, "permission denied"),
		strings.Contains(lower, "access denied"),
		strings.Contains(lower, "access is denied"),
		strings.Contains(lower, "libusb_error_access"):
		status.permissionError = true
	case strings.Contains(lower, "device not found"),
		strings.Contains(lower, "failed to open device after"):
		status.deviceOpenError = true
	}
}

func outputLogLevel(line string) string {
	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "failed"),
		strings.Contains(lower, "error"),
		strings.Contains(lower, "invalid"),
		strings.Contains(lower, "mismatch"),
		strings.Contains(lower, "cannot open"),
		strings.Contains(lower, "unsupported"):
		return "error"
	case strings.Contains(lower, "warning"),
		strings.Contains(lower, "potentially dangerous"),
		strings.Contains(lower, "skipped"):
		return "warn"
	case strings.Contains(lower, "successfully"),
		strings.Contains(lower, "verified"):
		return "success"
	default:
		return "info"
	}
}

func (a *App) flashResult(waitErr error, status outputStatus) (*FlashResult, error) {
	a.emitLog("info", "═══════════════════════════════════════")

	if status.permissionError {
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

	if status.deviceOpenError {
		a.emitLog("error", "✗ DEVICE OPEN FAILED")
		a.emitLog("error", "═══════════════════════════════════════")
		return nil, newAppError(errCodeFlash, fmt.Errorf("device could not be opened"))
	}

	if waitErr == nil && status.successDetected {
		a.emitLog("success", "✓ FLASH COMPLETED SUCCESSFULLY")
		a.emitLog("success", "═══════════════════════════════════════")
		a.emitLog("info", "Device will reboot automatically")
		a.emitLog("info", "You can now disconnect your keyboard")

		return &FlashResult{
			Success: true,
			Message: "Device successfully flashed!",
		}, nil
	}

	if waitErr != nil {
		a.emitLog("error", "✗ FLASH FAILED")
		a.emitLog("error", fmt.Sprintf("Exit error: %v", waitErr))
		a.emitLog("error", "═══════════════════════════════════════")
		return nil, newAppError(errCodeFlash, waitErr)
	}

	a.emitLog("warn", "⚠ FLASH STATUS UNKNOWN")
	a.emitLog("warn", "Process completed but success not confirmed")
	a.emitLog("warn", "═══════════════════════════════════════")

	return nil, newAppError(errCodeFlash, fmt.Errorf("flash status could not be confirmed"))
}
