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
	if strings.TrimSpace(device.Path) == "" {
		return fmt.Errorf("device path is required")
	}

	return validateVIDPID(device.VID, device.PID)
}

func validateVIDPID(vid, pid string) error {
	_, _, err := parseVIDPID(vid, pid)
	return err
}

func parseVIDPID(vid, pid string) (uint16, uint16, error) {
	values := []struct {
		label string
		value string
	}{
		{label: "VID", value: vid},
		{label: "PID", value: pid},
	}

	parsed := make([]uint16, len(values))
	for i, item := range values {
		if len(item.value) != 4 {
			return 0, 0, fmt.Errorf("%s must be a four-character hexadecimal value", item.label)
		}

		value, err := strconv.ParseUint(item.value, 16, 16)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid %s %q: %w", item.label, item.value, err)
		}
		parsed[i] = uint16(value)
	}

	return parsed[0], parsed[1], nil
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
