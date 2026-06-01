//go:build windows

package main

import _ "embed"

//go:embed build/tray_icon.ico
var trayIconData []byte
