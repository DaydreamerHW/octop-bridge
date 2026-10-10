package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	BaseURL        string `json:"base_url"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
	Token          string `json:"token,omitempty"`
	DefaultAgentID string `json:"default_agent_id,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	Insecure       bool   `json:"insecure,omitempty"`
}

type TokenCacheEntry struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CachedAt  time.Time `json:"cached_at"`
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "octop-bridge"), nil
}

func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func CacheFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".cache", "octop-bridge", "token_cache.json"), nil
}

func (c *Config) HTTPURL() string {
	raw := strings.TrimRight(c.BaseURL, "/")
	if raw == "" {
		raw = "http://192.168.5.192:8089"
	}
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw
	}
	return raw
}

func (c *Config) WSURL() string {
	httpURL := c.HTTPURL()
	parsed, err := url.Parse(httpURL)
	if err != nil {
		return "ws://192.168.5.192:8089"
	}
	scheme := "ws"
	if parsed.Scheme == "https" {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s", scheme, parsed.Host)
}

func LoadFileConfig() (*Config, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveFileConfig(cfg *Config) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadCachedToken(baseURL string) string {
	cachePath, err := CacheFilePath()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return ""
	}
	var cache map[string]TokenCacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return ""
	}
	entry, ok := cache[baseURL]
	if !ok {
		return ""
	}
	// Buffer 60 seconds
	if time.Now().Add(60 * time.Second).Before(entry.ExpiresAt) {
		return entry.Token
	}
	return ""
}

func SaveCachedToken(baseURL, token string, expiresInSeconds int) {
	cachePath, err := CacheFilePath()
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(cachePath), 0700)

	cache := make(map[string]TokenCacheEntry)
	if data, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(data, &cache)
	}

	cache[baseURL] = TokenCacheEntry{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiresInSeconds) * time.Second),
		CachedAt:  time.Now(),
	}

	if data, err := json.MarshalIndent(cache, "", "  "); err == nil {
		_ = os.WriteFile(cachePath, data, 0600)
	}
}

func Resolve(urlFlag, usernameFlag, passwordFlag, tokenFlag, agentFlag string, insecureFlag bool, timeoutFlag int) *Config {
	fileCfg, _ := LoadFileConfig()
	if fileCfg == nil {
		fileCfg = &Config{}
	}

	// 1. Base URL
	targetURL := urlFlag
	if targetURL == "" {
		targetURL = os.Getenv("OCTOP_URL")
	}
	if targetURL == "" && os.Getenv("OCTOP_HOST") != "" {
		port := os.Getenv("OCTOP_PORT")
		if port == "" {
			port = "8089"
		}
		targetURL = fmt.Sprintf("http://%s:%s", os.Getenv("OCTOP_HOST"), port)
	}
	if targetURL == "" {
		targetURL = fileCfg.BaseURL
	}
	if targetURL == "" {
		targetURL = "http://192.168.5.192:8089"
	}

	// 2. Token
	resolvedToken := tokenFlag
	if resolvedToken == "" {
		resolvedToken = os.Getenv("OCTOP_TOKEN")
	}
	if resolvedToken == "" {
		resolvedToken = fileCfg.Token
	}

	// 3. Username / Password
	resolvedUser := usernameFlag
	if resolvedUser == "" {
		resolvedUser = os.Getenv("OCTOP_USERNAME")
	}
	if resolvedUser == "" {
		resolvedUser = fileCfg.Username
	}
	if resolvedUser == "" {
		resolvedUser = "xqx"
	}

	resolvedPass := passwordFlag
	if resolvedPass == "" {
		resolvedPass = os.Getenv("OCTOP_PASSWORD")
	}
	if resolvedPass == "" {
		resolvedPass = fileCfg.Password
	}
	if resolvedPass == "" {
		resolvedPass = "Xqxq@1102@2026"
	}

	// 4. Default Agent ID
	resolvedAgent := agentFlag
	if resolvedAgent == "" {
		resolvedAgent = os.Getenv("OCTOP_AGENT_ID")
	}
	if resolvedAgent == "" {
		resolvedAgent = fileCfg.DefaultAgentID
	}
	if resolvedAgent == "" {
		resolvedAgent = "PQG34P"
	}

	// 5. Timeout
	resolvedTimeout := timeoutFlag
	if resolvedTimeout <= 0 {
		if envTimeout := os.Getenv("OCTOP_TIMEOUT"); envTimeout != "" {
			var parsed int
			if _, err := fmt.Sscanf(envTimeout, "%d", &parsed); err == nil && parsed > 0 {
				resolvedTimeout = parsed
			}
		}
	}
	if resolvedTimeout <= 0 {
		if fileCfg.Timeout > 0 {
			resolvedTimeout = fileCfg.Timeout
		} else {
			resolvedTimeout = 180
		}
	}

	// 6. Insecure
	resolvedInsecure := insecureFlag || fileCfg.Insecure || strings.ToLower(os.Getenv("OCTOP_INSECURE")) == "true" || os.Getenv("OCTOP_INSECURE") == "1"

	return &Config{
		BaseURL:        targetURL,
		Username:       resolvedUser,
		Password:       resolvedPass,
		Token:          resolvedToken,
		DefaultAgentID: resolvedAgent,
		Timeout:        resolvedTimeout,
		Insecure:       resolvedInsecure,
	}
}
