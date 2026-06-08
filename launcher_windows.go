package main

import (
	"fmt"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows"
)

func (a *App) launchInstallerAndQuit(installerPath string) error {
	verbPtr, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}
	pathPtr, err := syscall.UTF16PtrFromString(installerPath)
	if err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}
	if err := windows.ShellExecute(0, verbPtr, pathPtr, nil, nil, windows.SW_SHOWNORMAL); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}
	runtime.Quit(a.ctx)
	return nil
}
