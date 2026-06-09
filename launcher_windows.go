package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/minio/selfupdate"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) applyUpdateAndRelaunch(binaryPath string) error {
	f, err := os.Open(binaryPath)
	if err != nil {
		return fmt.Errorf("self-update: open binary: %w", err)
	}
	defer f.Close()

	if err := selfupdate.Apply(f, selfupdate.Options{}); err != nil {
		return fmt.Errorf("self-update: apply: %w", err)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("self-update: resolve executable: %w", err)
	}
	if err := exec.Command(exe).Start(); err != nil {
		return fmt.Errorf("self-update: relaunch: %w", err)
	}
	runtime.Quit(a.ctx)
	return nil
}
