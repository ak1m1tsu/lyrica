package main

import "errors"

func (a *App) launchInstallerAndQuit(_ string) error {
	return errors.New("auto-update install not supported on this platform")
}
