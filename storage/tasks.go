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

// Task represents a GTD task.
type Task struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Project  string `json:"project"`
	Priority string `json:"priority"` // H, M, L
	Deadline string `json:"deadline"`
	Status   string `json:"status"`   // TODO, DOING, DONE
}

func taskPath(dataDir string) string {
	return filepath.Join(dataDir, "tasks.json")
}

// LoadTasks reads the task list from the data directory.
func LoadTasks(dataDir string) ([]Task, error) {
	path := taskPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// SaveTasks writes the task list to the data directory.
func SaveTasks(dataDir string, tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(taskPath(dataDir), data, 0644)
}
