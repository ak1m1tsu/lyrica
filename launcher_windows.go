package main

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) launchInstallerAndQuit(installerPath string) error {
	cmd := exec.Command(installerPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}
	runtime.Quit(a.ctx)
	return nil
}
