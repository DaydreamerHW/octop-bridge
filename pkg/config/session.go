package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

var sessionMu sync.Mutex

// SessionBindingsMap maps CallerSessionID (e.g. HERMES_SESSION_ID) -> map[AgentID]ThreadID
type SessionBindingsMap map[string]map[string]string

func SessionBindingsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "octop-bridge", "session_bindings.json"), nil
}

func DetectCallerSessionID(explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, envVar := range []string{"HERMES_SESSION_ID", "HERMES_SESSION_KEY", "AGENT_SESSION_ID", "SESSION_ID"} {
		if val := os.Getenv(envVar); val != "" {
			return val
		}
	}
	return "cli-default"
}

func LoadSessionBindings() (SessionBindingsMap, error) {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	path, err := SessionBindingsFilePath()
	if err != nil {
		return make(SessionBindingsMap), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(SessionBindingsMap), nil
		}
		return make(SessionBindingsMap), err
	}

	var store SessionBindingsMap
	if err := json.Unmarshal(data, &store); err != nil {
		return make(SessionBindingsMap), nil
	}
	if store == nil {
		store = make(SessionBindingsMap)
	}
	return store, nil
}

func SaveSessionBindings(store SessionBindingsMap) error {
	sessionMu.Lock()
	defer sessionMu.Unlock()

	path, err := SessionBindingsFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func GetBoundThreadID(callerSessionID, agentID string) string {
	store, err := LoadSessionBindings()
	if err != nil || store == nil {
		return ""
	}
	agentMap, ok := store[callerSessionID]
	if !ok {
		return ""
	}
	return agentMap[agentID]
}

func SetBoundThreadID(callerSessionID, agentID, threadID string) error {
	store, _ := LoadSessionBindings()
	if store == nil {
		store = make(SessionBindingsMap)
	}

	if _, ok := store[callerSessionID]; !ok {
		store[callerSessionID] = make(map[string]string)
	}
	store[callerSessionID][agentID] = threadID
	return SaveSessionBindings(store)
}

func ClearBoundThreadID(callerSessionID, agentID string) error {
	store, _ := LoadSessionBindings()
	if store == nil {
		return nil
	}
	if agentMap, ok := store[callerSessionID]; ok {
		delete(agentMap, agentID)
		if len(agentMap) == 0 {
			delete(store, callerSessionID)
		}
		return SaveSessionBindings(store)
	}
	return nil
}
