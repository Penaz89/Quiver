package tui

import (
	"context"
	"sort"
	"sync"
)

var (
	sessionsMu sync.Mutex
	// activeSessions maps a connection context to a logged-in username
	activeSessions = make(map[context.Context]string)
)

// registerSessionLogin associates a username with a connection context
func registerSessionLogin(ctx context.Context, username string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	activeSessions[ctx] = username
}

// registerSessionLogout removes the username association but keeps the context
// tracking alive until the SSH connection drops
func registerSessionLogout(ctx context.Context) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(activeSessions, ctx)
}

// cleanupSession is called when the SSH connection completely drops
func cleanupSession(ctx context.Context) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	delete(activeSessions, ctx)
}

// getActiveUsersList returns a sorted slice of unique active users
func getActiveUsersList() []string {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	
	usersMap := make(map[string]bool)
	for _, u := range activeSessions {
		if u != "" {
			usersMap[u] = true
		}
	}
	
	var list []string
	for u := range usersMap {
		list = append(list, u)
	}
	
	sort.Strings(list)
	return list
}
