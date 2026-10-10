package config

import (
	"os"
	"testing"
)

func TestConfigURLs(t *testing.T) {
	cfg := &Config{BaseURL: "http://192.168.1.50:8089"}
	if cfg.HTTPURL() != "http://192.168.1.50:8089" {
		t.Fatalf("expected http://192.168.1.50:8089, got %s", cfg.HTTPURL())
	}
	if cfg.WSURL() != "ws://192.168.1.50:8089" {
		t.Fatalf("expected ws://192.168.1.50:8089, got %s", cfg.WSURL())
	}

	cfgHTTPS := &Config{BaseURL: "https://octop.example.com/"}
	if cfgHTTPS.HTTPURL() != "https://octop.example.com" {
		t.Fatalf("expected https://octop.example.com, got %s", cfgHTTPS.HTTPURL())
	}
	if cfgHTTPS.WSURL() != "wss://octop.example.com" {
		t.Fatalf("expected wss://octop.example.com, got %s", cfgHTTPS.WSURL())
	}
}

func TestResolveEnvOverrides(t *testing.T) {
	os.Setenv("OCTOP_URL", "https://octop.nas:9000")
	os.Setenv("OCTOP_USERNAME", "admin")
	os.Setenv("OCTOP_PASSWORD", "secret")
	os.Setenv("OCTOP_AGENT_ID", "agent_999")
	os.Setenv("OCTOP_INSECURE", "true")
	os.Setenv("OCTOP_TIMEOUT", "60")
	defer func() {
		os.Unsetenv("OCTOP_URL")
		os.Unsetenv("OCTOP_USERNAME")
		os.Unsetenv("OCTOP_PASSWORD")
		os.Unsetenv("OCTOP_AGENT_ID")
		os.Unsetenv("OCTOP_INSECURE")
		os.Unsetenv("OCTOP_TIMEOUT")
	}()

	cfg := Resolve("", "", "", "", "", false, 0)
	if cfg.HTTPURL() != "https://octop.nas:9000" {
		t.Fatalf("expected https://octop.nas:9000, got %s", cfg.HTTPURL())
	}
	if cfg.WSURL() != "wss://octop.nas:9000" {
		t.Fatalf("expected wss://octop.nas:9000, got %s", cfg.WSURL())
	}
	if cfg.Username != "admin" {
		t.Fatalf("expected admin, got %s", cfg.Username)
	}
	if cfg.Password != "secret" {
		t.Fatalf("expected secret, got %s", cfg.Password)
	}
	if cfg.DefaultAgentID != "agent_999" {
		t.Fatalf("expected agent_999, got %s", cfg.DefaultAgentID)
	}
	if !cfg.Insecure {
		t.Fatalf("expected insecure=true")
	}
	if cfg.Timeout != 60 {
		t.Fatalf("expected timeout=60, got %d", cfg.Timeout)
	}
}
