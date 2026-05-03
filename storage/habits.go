// Quiver - An SSH TUI Application
// Copyright (C) 2026  penaz
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Habit represents a habit to track.
type Habit struct {
	Name      string          `json:"name"`
	Completed map[string]bool `json:"completed"` // keys: "2006-01-02"
	CreatedAt string          `json:"created_at"`
}

func habitPath(dataDir string) string {
	return filepath.Join(dataDir, "habits.json")
}

// LoadHabits reads the habit list from the data directory.
func LoadHabits(dataDir string) ([]Habit, error) {
	path := habitPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Habit{}, nil
		}
		return nil, err
	}
	var habits []Habit
	if err := json.Unmarshal(data, &habits); err != nil {
		return nil, err
	}
	// ensure maps are initialized
	for i := range habits {
		if habits[i].Completed == nil {
			habits[i].Completed = make(map[string]bool)
		}
	}
	return habits, nil
}

// SaveHabits writes the habit list to the data directory.
func SaveHabits(dataDir string, habits []Habit) error {
	data, err := json.MarshalIndent(habits, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(habitPath(dataDir), data, 0644)
}
