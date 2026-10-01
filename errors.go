package main

import (
	"fmt"
	"strings"
)

const (
	errCodeNoDevice       = "NO_DEVICE"
	errCodeUSBPermission  = "USB_PERMISSION"
	errCodeHIDEnumeration = "HID_ENUMERATION"
	errCodeHIDInit        = "HID_INIT"
	errCodeFlash          = "FLASH"
	errCodeInvalidInput   = "INVALID_INPUT"
	errCodeBusy           = "BUSY"
	errCodeTimeout        = "TIMEOUT"
	errCodeUnavailable    = "UNAVAILABLE"
)

type appError struct {
	Code string
	Err  error
}

func (e *appError) Error() string {
	if e == nil {
		return errCodeUnavailable
	}
	if e.Err == nil {
		return e.Code
	}

	return fmt.Sprintf("%s: %v", e.Code, e.Err)
}

func (e *appError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func newAppError(code string, err error) error {
	return &appError{
		Code: code,
		Err:  err,
	}
}

func classifyHIDError(err error) string {
	if err == nil {
		return errCodeHIDEnumeration
	}

	message := strings.ToLower(err.Error())

	if strings.Contains(message, "permission denied") ||
		strings.Contains(message, "access denied") {
		return errCodeUSBPermission
	}

	return errCodeHIDEnumeration
}
