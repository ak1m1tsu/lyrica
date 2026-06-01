package main

import (
	"log/slog"
	"sync"
	"time"

	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1511118352637231204"

type discordPresence struct {
	mu        sync.Mutex
	connected bool
	start     time.Time
}

func newDiscordPresence() *discordPresence {
	return &discordPresence{}
}

func (d *discordPresence) connect() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.connected {
		return
	}
	if err := client.Login(discordClientID); err != nil {
		slog.Warn("discord: connect failed", "error", err)
		return
	}
	d.connected = true
	d.start = time.Now()
}

func (d *discordPresence) timestamps() *client.Timestamps {
	return &client.Timestamps{Start: &d.start}
}

func (d *discordPresence) disconnect() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.connected {
		return
	}
	client.Logout()
	d.connected = false
}

func (d *discordPresence) setIdle() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.connected {
		return
	}
	if err := client.SetActivity(client.Activity{
		Details:    "Browsing lyrics",
		LargeImage: "logo",
		LargeText:  "Lyrica",
		Timestamps: d.timestamps(),
	}); err != nil {
		slog.Warn("discord: setIdle failed", "error", err)
	}
}

func (d *discordPresence) setSearching(query string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.connected {
		return
	}
	if len(query) > 128 {
		query = query[:128]
	}
	if err := client.SetActivity(client.Activity{
		Details:    "Searching",
		State:      query,
		LargeImage: "logo",
		LargeText:  "Lyrica",
		Timestamps: d.timestamps(),
	}); err != nil {
		slog.Warn("discord: setSearching failed", "error", err)
	}
}

func (d *discordPresence) setTrack(trackName, artistName string, synced bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.connected {
		return
	}
	smallImage := "plain"
	smallText := "Plain lyrics"
	if synced {
		smallImage = "synced"
		smallText = "Synced lyrics"
	}
	if err := client.SetActivity(client.Activity{
		Details:    trackName,
		State:      artistName,
		LargeImage: "logo",
		LargeText:  "Lyrica",
		SmallImage: smallImage,
		SmallText:  smallText,
		Timestamps: d.timestamps(),
	}); err != nil {
		slog.Warn("discord: setTrack failed", "error", err)
	}
}
