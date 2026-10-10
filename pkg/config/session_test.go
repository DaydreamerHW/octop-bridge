package config

import (
	"os"
	"testing"
)

func TestDetectCallerSessionID(t *testing.T) {
	// 1. Explicit takes precedence
	id := DetectCallerSessionID("custom-session-123")
	if id != "custom-session-123" {
		t.Fatalf("expected custom-session-123, got %s", id)
	}

	// 2. Env variable HERMES_SESSION_ID
	os.Setenv("HERMES_SESSION_ID", "hermes_session_999")
	defer os.Unsetenv("HERMES_SESSION_ID")

	id2 := DetectCallerSessionID("")
	if id2 != "hermes_session_999" {
		t.Fatalf("expected hermes_session_999, got %s", id2)
	}
}

func TestSessionBindingsMap(t *testing.T) {
	callerID := "test_caller_session_xyz"
	agentID := "agent_foo"
	threadID := "thr_bar_123"

	// Set binding
	err := SetBoundThreadID(callerID, agentID, threadID)
	if err != nil {
		t.Fatalf("failed to set bound thread: %v", err)
	}

	// Retrieve binding
	got := GetBoundThreadID(callerID, agentID)
	if got != threadID {
		t.Fatalf("expected %s, got %s", threadID, got)
	}

	// Clear binding
	err = ClearBoundThreadID(callerID, agentID)
	if err != nil {
		t.Fatalf("failed to clear bound thread: %v", err)
	}

	// Verify cleared
	gotAfter := GetBoundThreadID(callerID, agentID)
	if gotAfter != "" {
		t.Fatalf("expected empty thread after clear, got %s", gotAfter)
	}
}
