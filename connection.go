package main

import (
	"fmt"
	"runtime"

	"github.com/sstallion/go-hid"
)

// resolveConnectedDevice confirms that the same HID path detected by the UI is
// still connected immediately before a destructive flash operation. It rebuilds
// transport-derived fields from the current HID record instead of trusting them
// from the Wails request.
func resolveConnectedDevice(device *Device) (*Device, error) {
	if err := validateDevice(device); err != nil {
		return nil, err
	}

	vid, pid, err := parseVIDPID(device.VID, device.PID)
	if err != nil {
		return nil, err
	}

	var connected *hid.DeviceInfo
	err = hid.Enumerate(vid, pid, func(info *hid.DeviceInfo) error {
		if info == nil || info.Path != device.Path {
			return nil
		}

		copy := *info
		connected = &copy
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("re-enumerate connected device: %w", err)
	}
	if connected == nil {
		return nil, fmt.Errorf("detected device is no longer connected")
	}

	isBootloader := connected.VendorID == SONIX_VID
	if isBootloader {
		if _, ok := knownBootloaderPIDs[connected.ProductID]; !ok {
			return nil, fmt.Errorf("connected bootloader is not supported")
		}
	} else if !knownApplicationDevice(connected) {
		return nil, fmt.Errorf("connected application device is not supported")
	}

	if device.IsBootloader != isBootloader {
		return nil, fmt.Errorf("device mode changed; detect the device again")
	}
	if !isBootloader && !knownApplicationFirmware(connected, device.FirmwarePath) {
		return nil, fmt.Errorf("selected firmware does not match the connected device")
	}

	if runtime.GOOS == "linux" {
		hidDevice, err := hid.OpenPath(connected.Path)
		if err != nil {
			return nil, newAppError(errCodeUSBPermission, err)
		}
		if err := hidDevice.Close(); err != nil {
			return nil, fmt.Errorf("close USB access check: %w", err)
		}
	}

	resolved := *device
	resolved.VID = fmt.Sprintf("%04x", connected.VendorID)
	resolved.PID = fmt.Sprintf("%04x", connected.ProductID)
	resolved.Manufacturer = connected.MfrStr
	resolved.Product = connected.ProductStr
	resolved.SerialNumber = connected.SerialNbr
	resolved.Path = connected.Path
	resolved.IsBootloader = isBootloader

	return &resolved, nil
}

func knownApplicationDevice(info *hid.DeviceInfo) bool {
	if info == nil || info.VendorID != DESIGNEDBYGG_VID {
		return false
	}

	for _, device := range knownAppModePIDs[info.ProductID] {
		if device.BcdDevice == info.ReleaseNbr {
			return true
		}
	}

	return false
}

func knownApplicationFirmware(info *hid.DeviceInfo, firmwarePath string) bool {
	if info == nil || firmwarePath == "" {
		return false
	}

	for _, device := range knownAppModePIDs[info.ProductID] {
		if device.BcdDevice == info.ReleaseNbr && device.FirmwarePath == firmwarePath {
			return true
		}
	}

	return false
}
