package main

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"
	"encoding/json"
	"github.com/sstallion/go-hid"
)

const SONIX_VID = 0x0C45
const DESIGNEDBYGG_VID = 0x320F

type Device struct {
	VID          string `json:"vid"`
	PID          string `json:"pid"`
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	Product      string `json:"product"`
	SerialNumber string `json:"serialNumber"`
	Path         string `json:"path"`
	IsBootloader bool   `json:"isBootloader"`
	FirmwarePath string `json:"firmwarePath"`
	Candidates   string `json:"candidates,omitempty"`
}

type AppModeDevice struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	FirmwarePath  string `json:"firmwarePath"`
	BootloaderVID uint16 `json:"bootloaderVID"`
	BootloaderPID uint16 `json:"bootloaderPID"`
	BcdDevice     uint16 `json:"bcdDevice"`
}

var knownBootloaderPIDs = map[uint16]string{
	0x7900: "SN32F22x/23x/24x Bootloader",
	0x7040: "SN32F24xB Bootloader",
	0x7160: "SN32F24xC Bootloader",
	0x7010: "SN32F26x Bootloader",
	0x7120: "SN32F28x Bootloader",
	0x7140: "SN32F29x Bootloader",
}

var knownAppModePIDs = map[uint16][]AppModeDevice{
	0x5041: {
		{
			Name:          "BSK01",
			Description:   "DesignedbyGG Berserker",
			FirmwarePath:  "firmware/BSK01.bin",
			BootloaderVID: SONIX_VID,
			BootloaderPID: 0x7040,
			BcdDevice:     0x0103,
		},
		{
			Name:          "ICL01",
			Description:   "DesignedbyGG Ironclad v1",
			FirmwarePath:  "firmware/ICL01.bin",
			BootloaderVID: SONIX_VID,
			BootloaderPID: 0x7040,
			BcdDevice:     0x0108,
		},
		{
			Name:          "CSB01",
			Description:   "DesignedbyGG Champions Bane",
			FirmwarePath:  "",
			BootloaderVID: SONIX_VID,
			BootloaderPID: 0x7040,
			BcdDevice:     0x0108,
		},

	},
	0x511E: {
		{
			Name:          "ICL03",
			Description:   "DesignedbyGG Ironclad v3",
			FirmwarePath:  "firmware/ICL03.bin",
			BootloaderVID: SONIX_VID,
			BootloaderPID: 0x7040,
			BcdDevice:     0x0102,
		},
	},
	0x5136: {
		{
			Name:          "RB01",
			Description:   "DesignedbyGG RedBladeFS",
			FirmwarePath:  "firmware/RB01.bin",
			BootloaderVID: SONIX_VID,
			BootloaderPID: 0x7040,
			BcdDevice:     0x0101,
		},
	},
}

func (a *App) DetectDevice() (*Device, error) {
	a.emitLog("info", "═══════════════════════════════════════")
	a.emitLog("info", "Starting USB HID device scan...")
	a.emitLog("info", "═══════════════════════════════════════")

	var detectedDevices []*Device
	seenDevices := make(map[string]bool)
	deviceCount := 0

	err := hid.Enumerate(0, 0, func(info *hid.DeviceInfo) error {
		deviceCount++
		
		deviceKey := fmt.Sprintf("%04x:%04x:%s", info.VendorID, info.ProductID, info.SerialNbr)
		
		if seenDevices[deviceKey] {
			return nil
		}
		
		// Check for bootloader mode
		if info.VendorID == SONIX_VID {
			if chipName, ok := knownBootloaderPIDs[info.ProductID]; ok {
				seenDevices[deviceKey] = true
				
				device := &Device{
					VID:          fmt.Sprintf("%04x", info.VendorID),
					PID:          fmt.Sprintf("%04x", info.ProductID),
					Name:         chipName,
					Manufacturer: info.MfrStr,
					Product:      info.ProductStr,
					SerialNumber: info.SerialNbr,
					Path:         info.Path,
					IsBootloader: true,
					FirmwarePath: "",
				}
				detectedDevices = append(detectedDevices, device)
				
				a.emitLog("success", fmt.Sprintf("✓ Found in BOOTLOADER mode: %s", chipName))
				a.emitLog("info", fmt.Sprintf("  VID: 0x%s | PID: 0x%s", device.VID, device.PID))
				
				if runtime.GOOS == "linux" {
					if !a.CheckUdevRules(info.VendorID, info.ProductID) {
						a.emitLog("warn", "  ⚠ Udev rules missing for this device")
					}
				}
			}
		}
		
		// Check for application mode
		if info.VendorID == DESIGNEDBYGG_VID {
			if appDevices, ok := knownAppModePIDs[info.ProductID]; ok {
				var matchedDevices []AppModeDevice
				for i := range appDevices {
					if appDevices[i].BcdDevice == info.ReleaseNbr {
						matchedDevices = append(matchedDevices, appDevices[i])
					}
				}

				if len(matchedDevices) > 0 {
					var matchedDevice *AppModeDevice
					var candidatesJSON string
					
					if len(matchedDevices) > 1 {
						// Multiple devices with same bcdDevice - needs disambiguation
						matchedDevice = &matchedDevices[0]
						candidatesData, _ := json.Marshal(matchedDevices)
						candidatesJSON = string(candidatesData)
					} else {
						matchedDevice = &matchedDevices[0]
					}

					seenDevices[deviceKey] = true
					
					firmwareExists := matchedDevice.FirmwarePath != ""
					if firmwareExists {
						_, err := binaries.ReadFile(matchedDevice.FirmwarePath)
						if err != nil {
							firmwareExists = false
						}
					}

					device := &Device{
						VID:          fmt.Sprintf("%04x", info.VendorID),
						PID:          fmt.Sprintf("%04x", info.ProductID),
						Name:         matchedDevice.Description,
						Manufacturer: info.MfrStr,
						Product:      info.ProductStr,
						SerialNumber: info.SerialNbr,
						Path:         info.Path,
						IsBootloader: false,
						FirmwarePath: matchedDevice.FirmwarePath,
						Candidates:   candidatesJSON,
					}
					detectedDevices = append(detectedDevices, device)
					
					if len(matchedDevices) > 1 {
						a.emitLog("warn", "⚠ Multiple models detected - disambiguation required")
					}
					
					a.emitLog("success", fmt.Sprintf("✓ Found in APPLICATION mode: %s", matchedDevice.Description))
					a.emitLog("info", fmt.Sprintf("  VID: 0x%s | PID: 0x%s | bcdDevice: %d.%02d", 
						device.VID, device.PID, info.ReleaseNbr>>8, info.ReleaseNbr&0xFF))
					a.emitLog("info", fmt.Sprintf("  Model: %s", matchedDevice.Name))
					
					if firmwareExists {
						a.emitLog("success", fmt.Sprintf("  ✓ Embedded firmware: %s", matchedDevice.FirmwarePath))
					} else {
						a.emitLog("warn", fmt.Sprintf("  ✗ Firmware not available"))
					}
					a.emitLog("warn", "  Note: Requires --reboot to enter bootloader mode")
					
					if runtime.GOOS == "linux" {
						if !a.CheckUdevRules(info.VendorID, info.ProductID) {
							a.emitLog("warn", "  ⚠ Udev rules missing for this device")
						}
					}
				}
			}
		}
		
		return nil
	})

	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to enumerate HID devices: %v", err))
		return nil, fmt.Errorf("failed to enumerate HID devices: %v", err)
	}

	a.emitLog("info", fmt.Sprintf("Scanned %d HID devices total", deviceCount))

	if len(detectedDevices) == 0 {
		a.emitLog("error", "═══════════════════════════════════════")
		a.emitLog("error", "✗ NO COMPATIBLE DEVICE DETECTED")
		a.emitLog("error", "═══════════════════════════════════════")
		a.emitLog("warn", "Troubleshooting tips:")
		a.emitLog("warn", "  1. For bootloader mode:")
		a.emitLog("warn", "     - Hold RESET while plugging in USB")
		a.emitLog("warn", "     - Or use bootloader key combo")
		a.emitLog("warn", "     - Or short the bootloader pins")
		a.emitLog("warn", "  2. For application mode:")
		a.emitLog("warn", "     - Ensure keyboard is connected normally")
		a.emitLog("warn", "     - Flasher will reboot to bootloader automatically")
		a.emitLog("warn", "  3. Check USB cable connection")
		a.emitLog("warn", "  4. Try a different USB port")
		
		if runtime.GOOS == "linux" {
			a.emitLog("warn", "  5. Check udev rules and permissions")
		}
		
		return nil, fmt.Errorf("no compatible device detected")
	}

	device := detectedDevices[0]
	
	a.emitLog("success", "═══════════════════════════════════════")
	a.emitLog("success", fmt.Sprintf("✓ DEVICE READY: %s", device.Name))
	a.emitLog("success", fmt.Sprintf("  VID: 0x%s | PID: 0x%s", device.VID, device.PID))
	if device.IsBootloader {
		a.emitLog("info", "  Mode: BOOTLOADER (ready to flash)")
	} else {
		a.emitLog("info", "  Mode: APPLICATION (will reboot to bootloader)")
	}
	a.emitLog("success", "═══════════════════════════════════════")

	if len(detectedDevices) > 1 {
		a.emitLog("warn", fmt.Sprintf("Note: %d compatible devices detected, using first one", len(detectedDevices)))
	}

	return device, nil
}

func (a *App) GetKeyboardImage(device *Device) string {
	if device == nil {
		a.emitLog("warn", "device is nil")
		return ""
	}

	var modelName string

	// Extract model name from FirmwarePath
	if device.FirmwarePath != "" {
		parts := strings.Split(device.FirmwarePath, "/")
		if len(parts) > 0 {
			fileName := parts[len(parts)-1]
			modelName = strings.TrimSuffix(fileName, ".bin")
		}
	}

	// If no firmware path, parse candidates JSON to find model name
	if modelName == "" && device.Candidates != "" {
		a.emitLog("info", fmt.Sprintf("No firmware path, checking candidates for device '%s'", device.Name))
		var candidates []AppModeDevice
		if err := json.Unmarshal([]byte(device.Candidates), &candidates); err == nil {
			a.emitLog("info", fmt.Sprintf("Found %d candidates", len(candidates)))
			// Find the candidate that matches the device name
			for _, candidate := range candidates {
				a.emitLog("info", fmt.Sprintf("  Candidate: %s (%s) - fw: '%s'", candidate.Name, candidate.Description, candidate.FirmwarePath))
				if candidate.Description == device.Name {
					modelName = candidate.Name
					a.emitLog("success", fmt.Sprintf("Matched device name to candidate '%s'", modelName))
					break
				}
			}
		} else {
			a.emitLog("error", fmt.Sprintf("Failed to parse candidates JSON: %v", err))
		}
	}

	// If still empty, return empty
	if modelName == "" {
		a.emitLog("warn", "Could not determine model name")
		return ""
	}

	// Map model name to image path
	imagePath := fmt.Sprintf("images/%s.jpg", modelName)

	data, err := binaries.ReadFile(imagePath)
	if err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to read image '%s': %v", imagePath, err))
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:image/jpeg;base64,%s", encoded)
}

func (a *App) GetAvailableKeyboards() []AppModeDevice {
	var keyboards []AppModeDevice

	for _, devices := range knownAppModePIDs {
		for _, device := range devices {
			keyboards = append(keyboards, device)
		}
	}

	return keyboards
}

func (a *App) GetAppIcon() string {
	if len(icon) == 0 {
		a.emitLog("warn", "App icon not embedded")
		return ""
	}

	encoded := base64.StdEncoding.EncodeToString(icon)
	return fmt.Sprintf("data:image/png;base64,%s", encoded)
}