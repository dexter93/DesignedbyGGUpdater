package main

import (
	"runtime"

	"github.com/sstallion/go-hid"
)

// CheckUdevRules checks if the current user has access to a specific device
func (a *App) CheckUdevRules(vid uint16, pid uint16) bool {
	if runtime.GOOS != "linux" {
		return true
	}

	var canOpen bool
	
	err := hid.Enumerate(vid, pid, func(info *hid.DeviceInfo) error {
		dev, err := hid.OpenPath(info.Path)
		if err == nil {
			canOpen = true
			dev.Close()
		}
		return nil
	})
	
	if err != nil {
		return false
	}
	
	return canOpen
}

// GetUdevRulesContent returns the udev rules as a string
func (a *App) GetUdevRulesContent() string {
	rules := `# Sonix Keyboard Flasher - udev rules
# Installation: sudo cp 50-sonix-keyboards.rules /etc/udev/rules.d/
#               sudo udevadm control --reload-rules && sudo udevadm trigger

# Sonix Bootloader Devices (VID: 0x0C45)
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7900", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7900", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7040", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7040", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7160", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7160", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7010", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7010", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7120", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7120", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7140", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0c45", ATTRS{idProduct}=="7140", MODE="0666", TAG+="uaccess"

# DesignedbyGG Keyboards - Application Mode (VID: 0x320F)
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="5041", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="5041", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="511e", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="511e", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="hidraw", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="5136", MODE="0666", TAG+="uaccess"
SUBSYSTEM=="usb", ATTRS{idVendor}=="320f", ATTRS{idProduct}=="5136", MODE="0666", TAG+="uaccess"
`
	return rules
}