package main

import (
	"context"
	"embed"
	"fmt"
	"sync"
	"time"

	"github.com/sstallion/go-hid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed binaries
//go:embed firmware
//go:embed images
var binaries embed.FS

type App struct {
	mu sync.Mutex

	ctx        context.Context
	hidInitErr error
	hidReady   bool
	isFlashing bool
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	if a == nil {
		return
	}

	a.mu.Lock()
	a.ctx = ctx
	a.mu.Unlock()

	if err := hid.Init(); err != nil {
		a.mu.Lock()
		a.hidInitErr = err
		a.hidReady = false
		a.mu.Unlock()

		a.emitLog("error", fmt.Sprintf("Failed to initialize HID: %v", err))
		return
	}

	a.mu.Lock()
	a.hidInitErr = nil
	a.hidReady = true
	a.mu.Unlock()

	a.emitLog("info", "HID library initialized")
	a.emitLog("info", "Application started")
	a.emitLog("info", "DesignedbyGG Keyboard Flasher")
}

func (a *App) shutdown(context.Context) {
	if a == nil {
		return
	}

	a.mu.Lock()
	ready := a.hidReady
	a.hidReady = false
	a.mu.Unlock()

	if ready {
		if err := hid.Exit(); err != nil {
			a.emitLog("warn", fmt.Sprintf("Failed to shut down HID: %v", err))
		}
	}
}

func (a *App) emitLog(level, message string) {
	if a == nil {
		return
	}

	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()

	// Logging must never make an ordinary error path fail or panic.
	if ctx == nil {
		return
	}

	wailsruntime.EventsEmit(ctx, "log", LogEntry{
		Timestamp: time.Now().Format("15:04:05.000"),
		Level:     level,
		Message:   message,
	})
}

func (a *App) requireHID() error {
	if a == nil {
		return newAppError(errCodeUnavailable, fmt.Errorf("application is unavailable"))
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.hidInitErr != nil {
		return newAppError(errCodeHIDInit, a.hidInitErr)
	}
	if !a.hidReady {
		return newAppError(errCodeHIDInit, fmt.Errorf("HID is not initialized"))
	}

	return nil
}

func (a *App) applicationContext() context.Context {
	if a == nil {
		return context.Background()
	}

	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()

	if ctx == nil {
		return context.Background()
	}

	return ctx
}

func (a *App) beginFlash() error {
	if a == nil {
		return newAppError(errCodeUnavailable, fmt.Errorf("application is unavailable"))
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isFlashing {
		return newAppError(errCodeBusy, fmt.Errorf("a flash operation is already in progress"))
	}

	a.isFlashing = true
	return nil
}

func (a *App) endFlash() {
	if a == nil {
		return
	}

	a.mu.Lock()
	a.isFlashing = false
	a.mu.Unlock()
}

func (a *App) GetVersion() string {
	return Version
}
