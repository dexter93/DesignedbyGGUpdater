package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
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

	a.emitLog("info", fmt.Sprintf("Extracting embedded binary: %s", binaryPath))
	data, err := binaries.ReadFile(binaryPath)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to read embedded binary: %v", err))
		return "", fmt.Errorf("failed to read embedded binary: %v", err)
	}

	tmpDir := os.TempDir()
	execName := "sonixflasher"
	if runtime.GOOS == "windows" {
		execName = "sonixflasher.exe"
	}
	tmpPath := filepath.Join(tmpDir, execName)

	a.emitLog("info", fmt.Sprintf("Writing binary to: %s", tmpPath))
	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to write binary: %v", err))
		return "", fmt.Errorf("failed to write binary: %v", err)
	}

	a.emitLog("success", "Binary extracted successfully")
	return tmpPath, nil
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
				return &FlashResult{Success: false, Message: err.Error()}, nil
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
			return &FlashResult{Success: false, Message: err.Error()}, nil
		}
		firmwarePath = tmpPath
		cleanupFirmware = true
		a.emitLog("info", fmt.Sprintf("Using embedded firmware: %s", device.FirmwarePath))
	} else {
		err := fmt.Errorf("no firmware specified")
		a.emitLog("error", err.Error())
		return &FlashResult{Success: false, Message: err.Error()}, nil
	}

	if cleanupFirmware {
		defer os.Remove(firmwarePath)
	}

	a.emitLog("info", fmt.Sprintf("Offset: 0x%x (%d bytes)", offset, offset))
	a.emitLog("info", "═══════════════════════════════════════")
	
	binPath, err := a.GetSonixFlasherPath()
	if err != nil {
		return &FlashResult{Success: false, Message: err.Error()}, nil
	}
	defer os.Remove(binPath)

	vidpid := fmt.Sprintf("%s/%s", device.VID, device.PID)
	args := []string{
		"--vidpid", vidpid,
		"--file", firmwarePath,
	}
	
	if !device.IsBootloader {
		args = append(args, "--reboot", "hfd")
		a.emitLog("warn", "Device in application mode, will reboot to bootloader...")
	}
	
	if offset > 0 {
		args = append(args, "--offset", fmt.Sprintf("0x%x", offset))
	}

	a.emitLog("info", fmt.Sprintf("Executing: %s %s", filepath.Base(binPath), strings.Join(args, " ")))
	
	cmd := exec.Command(binPath, args...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to create stdout pipe: %v", err))
		return &FlashResult{Success: false, Message: err.Error()}, nil
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to create stderr pipe: %v", err))
		return &FlashResult{Success: false, Message: err.Error()}, nil
	}

	a.emitLog("info", "Starting sonixflasher process...")
	if err := cmd.Start(); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to start process: %v", err))
		return &FlashResult{Success: false, Message: err.Error()}, nil
	}

	successDetected := false
	permissionError := false

	outputDone := make(chan bool, 2)
	go func() {
		defer func() { outputDone <- true }()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			
			if strings.Contains(line, "Device succesfully flashed") || 
			   strings.Contains(line, "Flash Verification Checksum: OK") {
				successDetected = true
			}
			
			if strings.Contains(line, "Could not open the device") ||
			   strings.Contains(line, "Device failed to open") ||
			   strings.Contains(line, "Permission denied") {
				permissionError = true
			}
			
			level := "info"
			if strings.Contains(strings.ToLower(line), "error") || 
			   strings.Contains(strings.ToLower(line), "fail") {
				level = "error"
			} else if strings.Contains(strings.ToLower(line), "warn") {
				level = "warn"
			} else if strings.Contains(line, "✓") || 
			          strings.Contains(strings.ToLower(line), "success") ||
			          strings.Contains(strings.ToLower(line), "detected") ||
			          strings.Contains(strings.ToLower(line), "done") ||
			          strings.Contains(strings.ToLower(line), "ok") {
				level = "success"
			}
			
			a.emitLog(level, line)
		}
	}()
	
	go func() {
		defer func() { outputDone <- true }()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			
			if strings.Contains(line, "Could not open") ||
			   strings.Contains(line, "Permission denied") ||
			   strings.Contains(line, "access denied") {
				permissionError = true
			}
			
			a.emitLog("error", line)
		}
	}()

	<-outputDone
	<-outputDone

	a.emitLog("info", "Waiting for flash process to complete...")
	err = cmd.Wait()
	
	a.emitLog("info", "═══════════════════════════════════════")
	
	if permissionError {
		a.emitLog("error", "✗ USB PERMISSION ERROR")
		a.emitLog("error", "═══════════════════════════════════════")
		if runtime.GOOS == "linux" {
			a.emitLog("error", "Udev rules are required for USB access")
			a.emitLog("warn", "Install udev rules and reload:")
			a.emitLog("warn", "  sudo tee /etc/udev/rules.d/50-sonix-keyboards.rules")
			a.emitLog("warn", "  sudo udevadm control --reload-rules")
			a.emitLog("warn", "  sudo udevadm trigger")
		}
		return &FlashResult{
			Success: false,
			Message: "USB_PERMISSION_ERROR",
		}, nil
	}
	
	if successDetected {
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
		return &FlashResult{
			Success: false,
			Message: fmt.Sprintf("Flash failed: %v", err),
		}, nil
	}
	
	a.emitLog("warn", "⚠ FLASH STATUS UNKNOWN")
	a.emitLog("warn", "Process completed but success not confirmed")
	a.emitLog("warn", "═══════════════════════════════════════")
	return &FlashResult{
		Success: false,
		Message: "Flash status unclear - check logs",
	}, nil
}