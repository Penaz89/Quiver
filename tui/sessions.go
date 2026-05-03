package tui

import (
	"context"
	"sort"
	"sync"
	"time"
)

var (
	sessionsMu sync.Mutex
	// activeSessions maps a connection context to a logged-in username
	activeSessions = make(map[context.Context]string)
	
	chatHistory []ChatMessage
	chatSubs    []chan ChatMessage
)

// ChatMessage represents a single message in the internal chat
type ChatMessage struct {
	Sender    string
	Text      string
	Timestamp time.Time
}

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

// broadcastChat sends a message to all connected clients
func broadcastChat(sender, text string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	
	msg := ChatMessage{
		Sender:    sender,
		Text:      text,
		Timestamp: time.Now(),
	}
	
	chatHistory = append(chatHistory, msg)
	if len(chatHistory) > 100 {
		chatHistory = chatHistory[1:] // Keep last 100 messages
	}
	
	for _, sub := range chatSubs {
		select {
		case sub <- msg:
		default:
		}
	}
}

// SubscribeChat returns a channel that receives live chat messages
func SubscribeChat() chan ChatMessage {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	ch := make(chan ChatMessage, 100)
	chatSubs = append(chatSubs, ch)
	return ch
}

// UnsubscribeChat removes a channel from the live broadcast
func UnsubscribeChat(ch chan ChatMessage) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	for i, sub := range chatSubs {
		if sub == ch {
			chatSubs = append(chatSubs[:i], chatSubs[i+1:]...)
			close(ch)
			break
		}
	}
}

// GetChatHistory returns a copy of the current chat history
func GetChatHistory() []ChatMessage {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	res := make([]ChatMessage, len(chatHistory))
	copy(res, chatHistory)
	return res
}
