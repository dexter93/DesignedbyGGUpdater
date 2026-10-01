package main

import "time"

const appName = "DesignedbyGG Updater"

var (
	flashOperationTimeout      = 2 * time.Minute
	outputScannerInitialBuffer = 64 * 1024
	outputScannerMaxBuffer     = 1024 * 1024
)
