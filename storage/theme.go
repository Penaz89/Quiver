package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ThemeColors holds color hex values or ANSI codes for UI elements
type ThemeColors struct {
	Name            string `json:"name"`
	Border          string `json:"border"`
	BorderFocus     string `json:"border_focus"`
	Logo            string `json:"logo"`
	Version         string `json:"version"`
	MenuNormal      string `json:"menu_normal"`
	MenuSelectedFg  string `json:"menu_selected_fg"`
	MenuSelectedBg  string `json:"menu_selected_bg"`
	MenuActiveDimFg string `json:"menu_active_dim_fg"`
	MenuActiveDimBg string `json:"menu_active_dim_bg"`
	Title           string `json:"title"`
	Subtitle        string `json:"subtitle"`
	Info            string `json:"info"`
	Highlight       string `json:"highlight"`
	Dim             string `json:"dim"`
	Status          string `json:"status"`
	HelpBar         string `json:"help_bar"`
	FieldBg         string `json:"field_bg"`
}

var BuiltinThemes = map[string]ThemeColors{
	"default": {
		Name:            "Default",
		Border:          "63",
		BorderFocus:     "205",
		Logo:            "205",
		Version:         "241",
		MenuNormal:      "252",
		MenuSelectedFg:  "255",
		MenuSelectedBg:  "63",
		MenuActiveDimFg: "252",
		MenuActiveDimBg: "237",
		Title:           "205",
		Subtitle:        "141",
		Info:            "252",
		Highlight:       "212",
		Dim:             "241",
		Status:          "42",
		HelpBar:         "241",
		FieldBg:         "236",
	},
	"catppuccin": {
		Name:            "Catppuccin",
		Border:          "#cba6f7", // Mauve
		BorderFocus:     "#f38ba8", // Red
		Logo:            "#f38ba8", // Red
		Version:         "#7f849c", // Overlay0
		MenuNormal:      "#cdd6f4", // Text
		MenuSelectedFg:  "#1e1e2e", // Base
		MenuSelectedBg:  "#cba6f7", // Mauve
		MenuActiveDimFg: "#cdd6f4", // Text
		MenuActiveDimBg: "#45475a", // Surface1
		Title:           "#89b4fa", // Blue
		Subtitle:        "#b4befe", // Lavender
		Info:            "#cdd6f4", // Text
		Highlight:       "#f9e2af", // Yellow
		Dim:             "#a6adc8", // Subtext0
		Status:          "#a6e3a1", // Green
		HelpBar:         "#9399b2", // Overlay2
		FieldBg:         "#313244", // Surface0
	},
	"nord": {
		Name:            "Nord",
		Border:          "#81A1C1", // nord9
		BorderFocus:     "#88C0D0", // nord8
		Logo:            "#88C0D0", // nord8
		Version:         "#4C566A", // nord3
		MenuNormal:      "#D8DEE9", // nord4
		MenuSelectedFg:  "#2E3440", // nord0
		MenuSelectedBg:  "#81A1C1", // nord9
		MenuActiveDimFg: "#D8DEE9", // nord4
		MenuActiveDimBg: "#4C566A", // nord3
		Title:           "#8FBCBB", // nord7
		Subtitle:        "#B48EAD", // nord15
		Info:            "#E5E9F0", // nord5
		Highlight:       "#EBCB8B", // nord13
		Dim:             "#4C566A", // nord3
		Status:          "#A3BE8C", // nord14
		HelpBar:         "#4C566A", // nord3
		FieldBg:         "#3B4252", // nord1
	},
	"gruvbox": {
		Name:            "Gruvbox",
		Border:          "#a89984", // fg4
		BorderFocus:     "#d79921", // yellow
		Logo:            "#fe8019", // orange
		Version:         "#928374", // gray
		MenuNormal:      "#ebdbb2", // fg
		MenuSelectedFg:  "#282828", // bg
		MenuSelectedBg:  "#d79921", // yellow
		MenuActiveDimFg: "#ebdbb2", // fg
		MenuActiveDimBg: "#504945", // bg2
		Title:           "#fabd2f", // yellow
		Subtitle:        "#d3869b", // purple
		Info:            "#ebdbb2", // fg
		Highlight:       "#8ec07c", // aqua
		Dim:             "#928374", // gray
		Status:          "#b8bb26", // green
		HelpBar:         "#a89984", // fg4
		FieldBg:         "#3c3836", // bg1
	},
	"kanagawa": {
		Name:            "Kanagawa",
		Border:          "#54546d", // sumiInk6
		BorderFocus:     "#7e9cd8", // crystalBlue
		Logo:            "#ffa066", // surimiOrange
		Version:         "#727169", // fujiGray
		MenuNormal:      "#dcd7ba", // fujiWhite
		MenuSelectedFg:  "#1f1f28", // sumiInk1
		MenuSelectedBg:  "#7e9cd8", // crystalBlue
		MenuActiveDimFg: "#dcd7ba", // fujiWhite
		MenuActiveDimBg: "#2a2a37", // sumiInk3
		Title:           "#e6c384", // carpYellow
		Subtitle:        "#957fb8", // oniViolet
		Info:            "#dcd7ba", // fujiWhite
		Highlight:       "#7FB4CA", // springBlue
		Dim:             "#727169", // fujiGray
		Status:          "#76946a", // autumnGreen
		HelpBar:         "#727169", // fujiGray
		FieldBg:         "#363646", // sumiInk4
	},
	"zenbones": {
		Name:            "Zenbones",
		Border:          "#c4cacb",
		BorderFocus:     "#286486", // blue
		Logo:            "#a8334c", // red
		Version:         "#818f96", // dim
		MenuNormal:      "#2c363c", // fg
		MenuSelectedFg:  "#f0edec", // bg
		MenuSelectedBg:  "#286486", // blue
		MenuActiveDimFg: "#2c363c", // fg
		MenuActiveDimBg: "#cbd3d4",
		Title:           "#944927", // wood
		Subtitle:        "#88507d", // purple
		Info:            "#2c363c", // fg
		Highlight:       "#286486", // blue
		Dim:             "#818f96", // dim
		Status:          "#4f6c31", // green
		HelpBar:         "#818f96", // dim
		FieldBg:         "#e2dedd",
	},
	"everforest": {
		Name:            "Everforest",
		Border:          "#4a555b", // bg3
		BorderFocus:     "#a7c080", // green
		Logo:            "#e67e80", // red
		Version:         "#859289", // grey1
		MenuNormal:      "#d3c6aa", // fg
		MenuSelectedFg:  "#2b3339", // bg0
		MenuSelectedBg:  "#a7c080", // green
		MenuActiveDimFg: "#d3c6aa", // fg
		MenuActiveDimBg: "#3a454a", // bg2
		Title:           "#dbbc7f", // yellow
		Subtitle:        "#d699b6", // purple
		Info:            "#d3c6aa", // fg
		Highlight:       "#7fbbb3", // blue
		Dim:             "#859289", // grey1
		Status:          "#a7c080", // green
		HelpBar:         "#859289", // grey1
		FieldBg:         "#323c41", // bg1
	},
	"retro-green": {
		Name:            "Retro Green",
		Border:          "#1b8000", // dim green
		BorderFocus:     "#33ff00", // bright green
		Logo:            "#33ff00", // bright green
		Version:         "#1b8000", // dim green
		MenuNormal:      "#1b8000", // dim green
		MenuSelectedFg:  "#000000", // black
		MenuSelectedBg:  "#33ff00", // bright green
		MenuActiveDimFg: "#33ff00", // bright green
		MenuActiveDimBg: "#0a3300", // dark green
		Title:           "#33ff00", // bright green
		Subtitle:        "#1b8000", // dim green
		Info:            "#33ff00", // bright green
		Highlight:       "#ccffaa", // whitish green
		Dim:             "#1b8000", // dim green
		Status:          "#33ff00", // bright green
		HelpBar:         "#1b8000", // dim green
		FieldBg:         "#0a3300", // dark green
	},
}

func LoadTheme(dataDir string, themeID string) ThemeColors {
	if t, ok := BuiltinThemes[themeID]; ok {
		return t
	}

	// Try loading custom theme
	path := filepath.Join(dataDir, "themes", themeID+".json")
	data, err := os.ReadFile(path)
	if err == nil {
		var t ThemeColors
		if err := json.Unmarshal(data, &t); err == nil {
			return t
		}
	}

	return BuiltinThemes["default"]
}

func GetAvailableThemes(dataDir string) []string {
	themes := []string{"default", "catppuccin", "nord", "gruvbox", "kanagawa", "zenbones", "everforest", "retro-green"}

	// Add custom themes if present
	entries, err := os.ReadDir(filepath.Join(dataDir, "themes"))
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
				name := strings.TrimSuffix(e.Name(), ".json")
				// Avoid duplicates if a file has the same name as built-in
				isBuiltin := false
				for _, bt := range themes {
					if name == bt {
						isBuiltin = true
						break
					}
				}
				if !isBuiltin {
					themes = append(themes, name)
				}
			}
		}
	}
	return themes
}

func GetThemeName(dataDir string, themeID string) string {
	if t, ok := BuiltinThemes[themeID]; ok {
		return t.Name
	}

	path := filepath.Join(dataDir, "themes", themeID+".json")
	data, err := os.ReadFile(path)
	if err == nil {
		var t ThemeColors
		if err := json.Unmarshal(data, &t); err == nil {
			if t.Name != "" {
				return t.Name
			}
		}
	}

	return themeID
}
