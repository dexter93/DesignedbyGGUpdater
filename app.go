package main

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/sstallion/go-hid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed binaries
//go:embed firmware
//go:embed images
var binaries embed.FS

type App struct {
	ctx context.Context
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"` // info, success, error, warn
	Message   string `json:"message"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	if err := hid.Init(); err != nil {
		a.emitLog("error", fmt.Sprintf("Failed to initialize HID: %v", err))
	} else {
		a.emitLog("info", "HID library initialized")
	}
	
	a.emitLog("info", "Application started")
	a.emitLog("info", "DesignedbyGG Keyboard Flasher")
}

func (a *App) emitLog(level, message string) {
	wailsruntime.EventsEmit(a.ctx, "log", LogEntry{
		Timestamp: time.Now().Format("15:04:05.000"),
		Level:     level,
		Message:   message,
	})
}