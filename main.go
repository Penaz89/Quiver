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

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"charm.land/log/v2"
	"charm.land/wish/v2"
	"charm.land/wish/v2/activeterm"
	"charm.land/wish/v2/bubbletea"
	"charm.land/wish/v2/logging"
	"github.com/charmbracelet/ssh"
	"github.com/penaz/quiver/tui"
)

const (
	defaultHost    = "0.0.0.0"
	defaultPort    = "2222"
	defaultDataDir = "/data"
)

// Version information (set at build time via ldflags)
var (
	Version = "dev"
)

func main() {
	host := envOrDefault("QUIVER_HOST", defaultHost)
	port := envOrDefault("QUIVER_PORT", defaultPort)
	dataDir := envOrDefault("QUIVER_DATA_DIR", defaultDataDir)

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatal("Failed to create data directory", "path", dataDir, "error", err)
	}

	// Host key lives on the persistent volume so it survives container recreation
	hostKeyPath := filepath.Join(dataDir, "host_key_ed25519")

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		}),
		wish.WithPasswordAuth(func(ctx ssh.Context, password string) bool {
			return true
		}),
		wish.WithMiddleware(
			bubbletea.Middleware(func(s ssh.Session) (tui.Model, []tui.ProgramOption) {
				return tui.NewModel(s, dataDir, Version)
			}),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Fatal("Could not create server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Starting Quiver SSH server",
		"version", Version,
		"host", host,
		"port", port,
		"data_dir", dataDir,
	)
	fmt.Printf(`
  ╔═══════════════════════════════════════╗
  ║           🏹  Q U I V E R             ║
  ║       SSH TUI Application v%s         ║
  ╚═══════════════════════════════════════╝
  Listening on %s
  Data directory: %s

`, Version, net.JoinHostPort(host, port), dataDir)

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Server error", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Shutting down Quiver...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Shutdown error", "error", err)
	}
	log.Info("Quiver stopped. Goodbye! 👋")
}

// envOrDefault returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or empty.
func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
