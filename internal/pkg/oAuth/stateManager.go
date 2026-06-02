package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

const (
	cleanUpInterval = 3
	bufferSize      = 32
)

var (
	ErrStateNotFound = errors.New("state not found")
	ErrStateExpired  = errors.New("state expired")
)

type StateData struct {
	Flow     string
	UserID   string
	Provider string
}
type storedData struct {
	Data      StateData
	CreatedAt int64
}

type StateManager struct {
	states map[string]storedData
	mu     sync.RWMutex
	ttl    time.Duration
}

func NewStateManager(ttl time.Duration) *StateManager {
	sm := &StateManager{
		states: make(map[string]storedData),
		ttl:    ttl,
	}

	go sm.cleanup()

	return sm
}

func (sm *StateManager) Generate(stateData StateData) (string, error) {
	b := make([]byte, bufferSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)

	sm.mu.Lock()
	sm.states[state] = storedData{
		Data:      stateData,
		CreatedAt: time.Now().Unix(),
	}
	sm.mu.Unlock()

	return state, nil
}

func (sm *StateManager) Verify(state string) (StateData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	storedData, exists := sm.states[state]
	if !exists {
		return StateData{}, ErrStateNotFound
	}

	if time.Now().Unix()-storedData.CreatedAt > int64(sm.ttl.Seconds()) {
		delete(sm.states, state)
		return StateData{}, ErrStateExpired
	}

	delete(sm.states, state)
	return storedData.Data, nil
}

func (sm *StateManager) cleanup() {
	ticker := time.NewTicker(cleanUpInterval * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Unix()
		sm.mu.Lock()
		for state, storedData := range sm.states {
			if now-storedData.CreatedAt > int64(sm.ttl.Seconds()) {
				delete(sm.states, state)
			}
		}
		sm.mu.Unlock()
	}
}
