//go:build darwin

package main

import _ "embed"

//go:embed build/tray_icon.png
var trayIconData []byte
