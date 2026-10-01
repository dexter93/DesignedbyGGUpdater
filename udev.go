package main

import (
	"fmt"
	"runtime"

	"github.com/sstallion/go-hid"
)

type udevAccessStatus struct {
	CanOpen bool
	Found   bool
}

func checkUdevAccess(vid, pid uint16) (udevAccessStatus, error) {
	if runtime.GOOS != "linux" {
		return udevAccessStatus{CanOpen: true, Found: true}, nil
	}
	if vid == 0 || pid == 0 {
		return udevAccessStatus{}, fmt.Errorf("VID and PID must be non-zero")
	}

	var status udevAccessStatus

	err := hid.Enumerate(vid, pid, func(info *hid.DeviceInfo) error {
		if info == nil || info.Path == "" {
			return nil
		}

		status.Found = true

		dev, err := hid.OpenPath(info.Path)
		if err != nil {
			return nil
		}

		status.CanOpen = true
		return dev.Close()
	})
	if err != nil {
		return udevAccessStatus{}, err
	}

	return status, nil
}

func udevRulesContent() string {
	return `# Sonix Keyboard Flasher - udev rules
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
}

func (a *App) GetUdevRulesContent() string {
	return udevRulesContent()
}
