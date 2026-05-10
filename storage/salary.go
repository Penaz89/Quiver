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

// Salary represents an income entry.
type Salary struct {
	Year   string `json:"year"`
	Month  string `json:"month"`
	Gross  string `json:"gross"`
	Net     string `json:"net"`
	Account string `json:"account,omitempty"`
	Author  string `json:"author,omitempty"`
}

func salaryPath(dataDir string) string {
	return filepath.Join(dataDir, "salaries.json")
}

// LoadSalaries reads the salary list from the data directory.
func LoadSalaries(dataDir string) ([]Salary, error) {
	path := salaryPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Salary{}, nil
		}
		return nil, err
	}
	var salaries []Salary
	if err := json.Unmarshal(data, &salaries); err != nil {
		return nil, err
	}
	return salaries, nil
}

// SaveSalaries writes the salary list to the data directory.
func SaveSalaries(dataDir string, salaries []Salary) error {
	data, err := json.MarshalIndent(salaries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(salaryPath(dataDir), data, 0644)
}
