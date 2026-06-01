package main

import (
	"fyne.io/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// runTray starts the system tray and blocks until the tray exits.
// It must be called from a dedicated goroutine.
func (a *App) runTray() {
	systray.Run(a.onTrayReady, nil)
}

func (a *App) onTrayReady() {
	systray.SetIcon(trayIconData)
	systray.SetTitle("Lyrica")
	systray.SetTooltip("Lyrica")

	mToggle := systray.AddMenuItem("Show / Hide", "Toggle window visibility")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Lyrica")

	go func() {
		for {
			select {
			case <-mToggle.ClickedCh:
				if a.windowVisible {
					runtime.WindowHide(a.ctx)
					a.windowVisible = false
				} else {
					runtime.WindowShow(a.ctx)
					a.windowVisible = true
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				runtime.Quit(a.ctx)
				return
			}
		}
	}()
}
