package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Goal struct {
	Name     string `json:"name"`
	Target   string `json:"target"`
	Current  string `json:"current"`
	Deadline string `json:"deadline"`
}

func goalsPath(dataDir string) string {
	return filepath.Join(dataDir, "goals.json")
}

func LoadGoals(dataDir string) ([]Goal, error) {
	path := goalsPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Goal{}, nil
		}
		return nil, err
	}

	var goals []Goal
	if err := json.Unmarshal(data, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

func SaveGoals(dataDir string, goals []Goal) error {
	path := goalsPath(dataDir)
	data, err := json.MarshalIndent(goals, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
