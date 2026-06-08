package domain

// ThemeColors holds the five user-configurable base colours for a theme.
type ThemeColors struct {
	Background  string `json:"background"`
	Surface     string `json:"surface"`
	Text        string `json:"text"`
	Accent      string `json:"accent"`
	AccentLight string `json:"accentLight"`
}

// Theme is a named colour theme that can be applied to the UI.
type Theme struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Colors ThemeColors `json:"colors"`
}
