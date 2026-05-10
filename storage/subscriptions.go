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

// Subscription represents a recurring service subscription.
type Subscription struct {
	Service string `json:"service"`
	Cost    string `json:"cost"`
	Type    string `json:"type"` // "type.monthly" or "type.annual"
	Account string `json:"account,omitempty"`
}

func subscriptionsPath(dataDir string) string {
	return filepath.Join(dataDir, "subscriptions.json")
}

// LoadSubscriptions reads subscription records from the data directory.
func LoadSubscriptions(dataDir string) ([]Subscription, error) {
	path := subscriptionsPath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Subscription{}, nil // return empty slice if file doesn't exist yet
		}
		return nil, err
	}

	var subs []Subscription
	if err := json.Unmarshal(data, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

// SaveSubscriptions writes subscription records to the data directory.
func SaveSubscriptions(dataDir string, subs []Subscription) error {
	path := subscriptionsPath(dataDir)
	data, err := json.MarshalIndent(subs, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
