package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func validateDevice(device *Device) error {
	if device == nil {
		return fmt.Errorf("device is required")
	}

	return validateVIDPID(device.VID, device.PID)
}

func validateVIDPID(vid, pid string) error {
	for label, value := range map[string]string{"VID": vid, "PID": pid} {
		if len(value) != 4 {
			return fmt.Errorf("%s must be a four-character hexadecimal value", label)
		}
		if _, err := strconv.ParseUint(value, 16, 16); err != nil {
			return fmt.Errorf("invalid %s %q: %w", label, value, err)
		}
	}

	return nil
}

func validateCustomFirmware(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("firmware path is required")
	}
	if !strings.EqualFold(filepath.Ext(path), ".bin") {
		return fmt.Errorf("firmware must be a .bin file")
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("access firmware: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("firmware must be a regular file")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open firmware: %w", err)
	}

	return file.Close()
}
