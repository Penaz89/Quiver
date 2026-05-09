// Quiver - An SSH TUI Application
// Copyright (C) 2026  penaz
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Settings holds user-configurable application settings.
type Settings struct {
	Language         string `json:"language"`          // "en" or "it"
	WeatherLoc       string `json:"weather_loc"`       // e.g. "Rome"
	Theme            string `json:"theme"`             // e.g. "default", "catppuccin", "nord"
	DefaultWorkspace string `json:"default_workspace"` // e.g. "Personal" or FamilyID
}

// DefaultSettings returns settings with default values.
func DefaultSettings() Settings {
	return Settings{
		Language:         "it",
		WeatherLoc:       "Rome",
		Theme:            "default",
		DefaultWorkspace: "Personal",
	}
}

func settingsPath(dataDir string) string {
	return filepath.Join(dataDir, "settings.json")
}

// LoadSettings reads settings from the data directory.
// Returns default settings if the file does not exist yet.
func LoadSettings(dataDir string) Settings {
	path := settingsPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultSettings()
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return DefaultSettings()
	}
	if s.Language == "" {
		s.Language = "en"
	}
	if s.Theme == "" {
		s.Theme = "default"
	}
	if s.DefaultWorkspace == "" {
		s.DefaultWorkspace = "Personal"
	}
	return s
}

// SaveSettings writes settings to the data directory.
func SaveSettings(dataDir string, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(dataDir), data, 0644)
}
