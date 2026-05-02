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
	"time"
)

// Vehicle represents a registered vehicle.
type Vehicle struct {
	Brand        string    `json:"brand"`
	Model        string    `json:"model"`
	LicensePlate string    `json:"license_plate"`
	Owner        string    `json:"owner"`
	RoadTax      time.Time `json:"road_tax"`
	RoadTaxCost  string    `json:"road_tax_cost"`
	NTC          time.Time `json:"ntc"`
	NTCCost      string    `json:"ntc_cost"`
}

func vehiclePath(dataDir string) string {
	return filepath.Join(dataDir, "vehicles.json")
}

// LoadVehicles reads the vehicle list from the data directory.
// Returns an empty slice if the file does not exist yet.
func LoadVehicles(dataDir string) ([]Vehicle, error) {
	path := vehiclePath(dataDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Vehicle{}, nil
		}
		return nil, err
	}
	var vehicles []Vehicle
	if err := json.Unmarshal(data, &vehicles); err != nil {
		return nil, err
	}
	return vehicles, nil
}

// SaveVehicles writes the vehicle list to the data directory.
func SaveVehicles(dataDir string, vehicles []Vehicle) error {
	data, err := json.MarshalIndent(vehicles, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(vehiclePath(dataDir), data, 0644)
}
