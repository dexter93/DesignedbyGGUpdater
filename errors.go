package main

import (
	"fmt"
	"strings"
)

const (
	errCodeNoDevice       = "NO_DEVICE"
	errCodeUSBPermission  = "USB_PERMISSION"
	errCodeHIDEnumeration = "HID_ENUMERATION"
	errCodeFlash          = "FLASH"
)

type appError struct {
	Code string
	Err  error
}

func (e *appError) Error() string {
	if e.Err == nil {
		return e.Code
	}

	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *appError) Unwrap() error {
	return e.Err
}

func newAppError(code string, err error) error {
	return &appError{
		Code: code,
		Err:  err,
	}
}

func classifyHIDError(err error) string {
	message := strings.ToLower(err.Error())

	if strings.Contains(message, "permission denied") ||
		strings.Contains(message, "access denied") {
		return errCodeUSBPermission
	}

	return errCodeHIDEnumeration
}
