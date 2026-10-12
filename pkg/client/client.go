package client

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DaydreamerHW/octop-bridge/pkg/config"
	"github.com/gorilla/websocket"
)

type OctopClient struct {
	Cfg        *config.Config
	HTTPClient *http.Client
}

func NewClient(cfg *config.Config) *OctopClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.Insecure,
		},
	}
	return &OctopClient{
		Cfg: cfg,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

func (c *OctopClient) Login(force bool) (string, error) {
	if !force && c.Cfg.Token != "" {
		return c.Cfg.Token, nil
	}

	if !force {
		cached := config.LoadCachedToken(c.Cfg.HTTPURL())
		if cached != "" {
			c.Cfg.Token = cached
			return cached, nil
		}
	}

	if c.Cfg.Username == "" || c.Cfg.Password == "" {
		return "", fmt.Errorf("username and password are required to login to Octop")
	}

	loginURL := fmt.Sprintf("%s/api/auth/login", c.Cfg.HTTPURL())
	body, _ := json.Marshal(map[string]string{
		"username": c.Cfg.Username,
		"password": c.Cfg.Password,
	})

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Octop at %s: %w", loginURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("octop login failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var res LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("invalid login response: %w", err)
	}
	if res.AccessToken == "" {
		return "", fmt.Errorf("login response did not contain access_token")
	}

	c.Cfg.Token = res.AccessToken
	config.SaveCachedToken(c.Cfg.HTTPURL(), res.AccessToken, res.ExpiresIn)
	return res.AccessToken, nil
}

func (c *OctopClient) apiGet(path string) ([]byte, error) {
	token, err := c.Login(false)
	if err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s%s", c.Cfg.HTTPURL(), path)
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Retry once with fresh token
		token, err = c.Login(true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API GET %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *OctopClient) apiPost(path string, payload interface{}) ([]byte, error) {
	token, err := c.Login(false)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	reqURL := fmt.Sprintf("%s%s", c.Cfg.HTTPURL(), path)
	req, err := http.NewRequest("POST", reqURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API POST %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *OctopClient) apiPatch(path string, payload interface{}) ([]byte, error) {
	token, err := c.Login(false)
	if err != nil {
		return nil, err
	}

	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	reqURL := fmt.Sprintf("%s%s", c.Cfg.HTTPURL(), path)
	req, err := http.NewRequest("PATCH", reqURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API PATCH %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *OctopClient) ListAgents() ([]AgentSummary, error) {
	raw, err := c.apiGet("/api/agents")
	if err != nil {
		return nil, err
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(raw, &list); err != nil {
		// Check if it's wrapped in an object
		var wrapped struct {
			Agents []map[string]interface{} `json:"agents"`
		}
		if err2 := json.Unmarshal(raw, &wrapped); err2 == nil {
			list = wrapped.Agents
		} else {
			return nil, err
		}
	}

	var summaries []AgentSummary
	for _, item := range list {
		aid, _ := item["agent_id"].(string)
		if aid == "" {
			aid, _ = item["id"].(string)
		}
		name, _ := item["name"].(string)
		desc, _ := item["description"].(string)
		model, _ := item["default_model"].(string)
		state, _ := item["state"].(string)
		kind, _ := item["kind"].(string)
		color, _ := item["color"].(string)
		icon, _ := item["icon_name"].(string)
		iconURL, _ := item["icon_url"].(string)
		shared, _ := item["is_shared"].(bool)

		enabled, ok := item["enabled"].(bool)
		if !ok {
			enabled = state != "stopped" && state != "disabled"
		}

		summaries = append(summaries, AgentSummary{
			AgentID:      aid,
			Name:         name,
			Description:  desc,
			DefaultModel: model,
			State:        state,
			Kind:         kind,
			Enabled:      enabled,
			IsShared:     shared,
			IconName:     icon,
			IconURL:      iconURL,
			Color:        color,
		})
	}
	return summaries, nil
}

func (c *OctopClient) ResolveAgent(target string) (*AgentSummary, error) {
	agents, err := c.ListAgents()
	if err != nil {
		return nil, err
	}

	ident := strings.TrimSpace(target)
	if ident == "" {
		ident = strings.TrimSpace(c.Cfg.DefaultAgentID)
	}

	if ident == "" {
		if len(agents) == 1 {
			return &agents[0], nil
		}
		for _, a := range agents {
			if strings.Contains(a.Name, "通用") || strings.Contains(strings.ToLower(a.Name), "assistant") {
				return &a, nil
			}
		}
		if len(agents) > 0 {
			return &agents[0], nil
		}
	}

	// 1. Exact ID
	for _, a := range agents {
		if a.AgentID == ident {
			return &a, nil
		}
	}

	// 2. Exact Name
	for _, a := range agents {
		if strings.EqualFold(a.Name, ident) {
			return &a, nil
		}
	}

	// 3. Substring Name
	for _, a := range agents {
		if strings.Contains(strings.ToLower(a.Name), strings.ToLower(ident)) {
			return &a, nil
		}
	}

	// 4. Try matching teams if not found among regular agents
	if teams, err := c.ListTeams(); err == nil {
		for _, t := range teams {
			if t.TeamID == ident || t.AgentID == ident || strings.EqualFold(t.Name, ident) || strings.Contains(strings.ToLower(t.Name), strings.ToLower(ident)) {
				return &AgentSummary{
					AgentID:      t.AgentID,
					Name:         t.Name,
					Description:  t.Description,
					DefaultModel: t.DefaultModel,
					State:        t.State,
					Kind:         "team",
					Enabled:      t.State != "stopped",
					IconName:     t.IconName,
					Color:        t.Color,
				}, nil
			}
		}
	}

	return &AgentSummary{AgentID: ident, Name: ident, Description: "Specified Agent"}, nil
}

func (c *OctopClient) GetAgentStatus(identifier string) (*AgentRuntimeStatus, error) {
	agent, err := c.ResolveAgent(identifier)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/agents/%s/status", agent.AgentID)
	raw, err := c.apiGet(path)
	if err != nil {
		return nil, err
	}
	var status AgentRuntimeStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

func (c *OctopClient) StartAgent(identifier string) error {
	agent, err := c.ResolveAgent(identifier)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/start", agent.AgentID)
	_, err = c.apiPost(path, nil)
	return err
}

func (c *OctopClient) StopAgent(identifier string) error {
	agent, err := c.ResolveAgent(identifier)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/stop", agent.AgentID)
	_, err = c.apiPost(path, nil)
	return err
}

func (c *OctopClient) ReloadAgent(identifier string) error {
	agent, err := c.ResolveAgent(identifier)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/reload", agent.AgentID)
	_, err = c.apiPost(path, nil)
	return err
}

func (c *OctopClient) DeleteAgent(identifier string) error {
	agent, err := c.ResolveAgent(identifier)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s", agent.AgentID)
	return c.apiDelete(path)
}

func (c *OctopClient) CreateThread(agentID string) (string, error) {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("/api/agents/%s/threads", agent.AgentID)
	raw, err := c.apiPost(path, map[string]interface{}{})
	if err != nil {
		return "", err
	}

	var res struct {
		ThreadID string `json:"thread_id"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	if res.ThreadID == "" {
		return "", fmt.Errorf("no thread_id in response: %s", string(raw))
	}
	return res.ThreadID, nil
}

func (c *OctopClient) ListThreads(agentID string) ([]ThreadSummary, error) {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/agents/%s/threads", agent.AgentID)
	raw, err := c.apiGet(path)
	if err != nil {
		return nil, err
	}

	var list []map[string]interface{}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}

	var summaries []ThreadSummary
	for _, item := range list {
		tid, _ := item["thread_id"].(string)
		title, _ := item["title"].(string)
		preview, _ := item["preview"].(string)
		created, _ := item["created_at"].(float64)
		updated, _ := item["updated_at"].(float64)

		summaries = append(summaries, ThreadSummary{
			ThreadID:  tid,
			Title:     title,
			Preview:   preview,
			CreatedAt: int64(created),
			UpdatedAt: int64(updated),
		})
	}
	return summaries, nil
}

func (c *OctopClient) apiDelete(path string) error {
	token, err := c.Login(false)
	if err != nil {
		return err
	}

	reqURL := fmt.Sprintf("%s%s", c.Cfg.HTTPURL(), path)
	req, err := http.NewRequest("DELETE", reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API DELETE %s failed (HTTP %d): %s", path, resp.StatusCode, string(body))
	}
	return nil
}

func (c *OctopClient) ListSkills(agentID string) ([]SkillSummary, error) {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/api/agents/%s/skills", agent.AgentID)
	raw, err := c.apiGet(path)
	if err != nil {
		return nil, err
	}

	var rawList []map[string]interface{}
	if err := json.Unmarshal(raw, &rawList); err != nil {
		return nil, err
	}

	var res []SkillSummary
	for _, item := range rawList {
		name, _ := item["name"].(string)
		desc, _ := item["description"].(string)
		enabled, _ := item["enabled"].(bool)
		kind, _ := item["kind"].(string)
		zh, _ := item["display_name_zh"].(string)
		en, _ := item["display_name_en"].(string)
		res = append(res, SkillSummary{
			Name:          name,
			Description:   desc,
			Enabled:       enabled,
			Kind:          kind,
			DisplayNameZH: zh,
			DisplayNameEN: en,
		})
	}
	return res, nil
}

func (c *OctopClient) EnableSkill(agentID, skillName string) error {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/skills/%s/enable", agent.AgentID, url.PathEscape(skillName))
	_, err = c.apiPost(path, nil)
	return err
}

func (c *OctopClient) DisableSkill(agentID, skillName string) error {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/skills/%s/disable", agent.AgentID, url.PathEscape(skillName))
	_, err = c.apiPost(path, nil)
	return err
}

func (c *OctopClient) DeleteSkill(agentID, skillName string) error {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return err
	}
	path := fmt.Sprintf("/api/agents/%s/skills/%s", agent.AgentID, url.PathEscape(skillName))
	return c.apiDelete(path)
}

func (c *OctopClient) InstallSkill(agentID, localPath string, overwrite bool) (*SkillSummary, error) {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("skill path not found: %w", err)
	}

	var skillName string
	var skillContent string
	var files []SkillFilePart

	if fi.IsDir() {
		// Read directory
		skillName = filepath.Base(filepath.Clean(localPath))
		skillMDPath := filepath.Join(localPath, "SKILL.md")
		mdBytes, err := os.ReadFile(skillMDPath)
		if err != nil {
			return nil, fmt.Errorf("SKILL.md missing in directory: %w", err)
		}
		skillContent = string(mdBytes)
		files = append(files, SkillFilePart{
			Path:          "SKILL.md",
			ContentBase64: base64.StdEncoding.EncodeToString(mdBytes),
		})

		err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			rel, err := filepath.Rel(localPath, path)
			if err != nil || rel == "SKILL.md" || strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files = append(files, SkillFilePart{
				Path:          rel,
				ContentBase64: base64.StdEncoding.EncodeToString(content),
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error reading skill files: %w", err)
		}
	} else if strings.HasSuffix(strings.ToLower(localPath), ".zip") {
		// Read ZIP file
		r, err := zip.OpenReader(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open zip file: %w", err)
		}
		defer r.Close()

		baseZipName := strings.TrimSuffix(filepath.Base(localPath), filepath.Ext(localPath))
		skillName = baseZipName

		for _, f := range r.File {
			if f.FileInfo().IsDir() {
				continue
			}
			name := filepath.Clean(f.Name)
			// Strip single parent directory if zipped as a folder
			parts := strings.Split(name, string(filepath.Separator))
			rel := name
			if len(parts) > 1 {
				skillName = parts[0]
				rel = strings.Join(parts[1:], "/")
			}

			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return nil, err
			}

			if rel == "SKILL.md" {
				skillContent = string(content)
			}
			files = append(files, SkillFilePart{
				Path:          rel,
				ContentBase64: base64.StdEncoding.EncodeToString(content),
			})
		}

		if skillContent == "" {
			return nil, fmt.Errorf("zip does not contain a SKILL.md file")
		}
	} else {
		return nil, fmt.Errorf("unsupported path: must be a skill directory or .zip file")
	}

	reqBody := CreateSkillRequest{
		Name:      skillName,
		Content:   skillContent,
		Files:     files,
		Overwrite: overwrite,
	}

	path := fmt.Sprintf("/api/agents/%s/skills", agent.AgentID)
	raw, err := c.apiPost(path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to install skill: %w", err)
	}

	// Auto-enable installed skill
	_ = c.EnableSkill(agent.AgentID, skillName)

	var res map[string]interface{}
	_ = json.Unmarshal(raw, &res)

	desc, _ := res["description"].(string)
	return &SkillSummary{
		Name:        skillName,
		Description: desc,
		Enabled:     true,
		Kind:        "workspace",
	}, nil
}

func (c *OctopClient) ChatStream(
	ctx context.Context,
	query string,
	agentID string,
	threadID string,
	newThread bool,
	callerSessionID string,
	mode string,
	onToken func(string),
	onToolStart func(string),
	onToolEnd func(string),
) (*TurnResult, error) {
	agent, err := c.ResolveAgent(agentID)
	if err != nil {
		return nil, err
	}

	sessionID := config.DetectCallerSessionID(callerSessionID)

	if newThread {
		// Explicit request to start a fresh clean session
		tid, err := c.CreateThread(agent.AgentID)
		if err != nil {
			return nil, fmt.Errorf("failed to create new thread: %w", err)
		}
		threadID = tid
		_ = config.SetBoundThreadID(sessionID, agent.AgentID, threadID)
	} else if threadID != "" {
		// Explicit thread requested by caller
		_ = config.SetBoundThreadID(sessionID, agent.AgentID, threadID)
	} else {
		// 1:1 Session Scoping: Check if current caller session already has a dedicated thread
		bound := config.GetBoundThreadID(sessionID, agent.AgentID)
		if bound != "" {
			threadID = bound
		} else {
			// First interaction in this session -> create a dedicated clean thread!
			tid, err := c.CreateThread(agent.AgentID)
			if err != nil {
				// Fallback to active thread if creation fails
				threadID = ""
			} else {
				threadID = tid
				_ = config.SetBoundThreadID(sessionID, agent.AgentID, threadID)
			}
		}
	}

	token, err := c.Login(false)
	if err != nil {
		return nil, err
	}

	wsURL := fmt.Sprintf("%s/api/agents/%s/chat/ws?token=%s", c.Cfg.WSURL(), agent.AgentID, token)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Cfg.Insecure,
		},
		HandshakeTimeout: 15 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial failed: %w", err)
	}
	defer conn.Close()

	if mode == "" {
		mode = "craft"
	}

	payload := InboundTurnPayload{
		Type:             "user_turn",
		Text:             query,
		ThreadID:         threadID,
		ConversationMode: mode,
	}

	if err := conn.WriteJSON(payload); err != nil {
		return nil, fmt.Errorf("failed to send turn payload: %w", err)
	}

	var fullResponse strings.Builder
	var toolsCalled []string
	finalThreadID := threadID
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}
			return nil, fmt.Errorf("error reading websocket message: %w", err)
		}

		var frame map[string]interface{}
		if err := json.Unmarshal(message, &frame); err != nil {
			continue
		}

		fType, _ := frame["type"].(string)

		if tid, ok := frame["thread_id"].(string); ok && tid != "" {
			finalThreadID = tid
		}

		switch fType {
		case "token":
			if content, ok := frame["content"].(string); ok && content != "" {
				fullResponse.WriteString(content)
				if onToken != nil {
					onToken(content)
				}
			}
		case "chunk", "text_delta":
			delta, _ := frame["delta"].(string)
			if delta == "" {
				delta, _ = frame["content"].(string)
			}
			if delta != "" {
				fullResponse.WriteString(delta)
				if onToken != nil {
					onToken(delta)
				}
			}
		case "tool_start", "tool_call":
			toolName, _ := frame["tool"].(string)
			if toolName == "" {
				toolName, _ = frame["name"].(string)
			}
			if toolName != "" {
				toolsCalled = append(toolsCalled, toolName)
				if onToolStart != nil {
					onToolStart(toolName)
				}
			}
		case "tool_end", "tool_result":
			toolName, _ := frame["tool"].(string)
			if toolName == "" {
				toolName, _ = frame["name"].(string)
			}
			if onToolEnd != nil {
				onToolEnd(toolName)
			}
		case "state_update":
			if data, ok := frame["data"].(map[string]interface{}); ok {
				if msgs, ok := data["messages"].([]interface{}); ok {
					for _, m := range msgs {
						if mObj, ok := m.(map[string]interface{}); ok {
							if tcs, ok := mObj["tool_calls"].([]interface{}); ok {
								for _, tc := range tcs {
									if tcObj, ok := tc.(map[string]interface{}); ok {
										fn, _ := tcObj["function"].(map[string]interface{})
										name, _ := fn["name"].(string)
										if name == "" {
											name, _ = tcObj["name"].(string)
										}
										if name != "" {
											toolsCalled = append(toolsCalled, name)
											if onToolStart != nil {
												onToolStart(name)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		case "done", "turn_done":
			goto Completed
		case "error":
			errMsg, _ := frame["message"].(string)
			return nil, fmt.Errorf("agent returned error: %s", errMsg)
		}
	}

Completed:
	if finalThreadID != "" {
		_ = config.SetBoundThreadID(sessionID, agent.AgentID, finalThreadID)
	}
	return &TurnResult{
		Status:          "success",
		AgentID:         agent.AgentID,
		AgentName:       agent.Name,
		ThreadID:        finalThreadID,
		Response:        strings.TrimSpace(fullResponse.String()),
		ToolsCalled:     toolsCalled,
		DurationSeconds: time.Since(startTime).Seconds(),
	}, nil
}

// --- Expert Templates API ---

func (c *OctopClient) ListExperts() ([]ExpertTemplate, error) {
	raw, err := c.apiGet("/api/experts")
	if err != nil {
		return nil, err
	}
	var list []ExpertTemplate
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *OctopClient) GetExpert(id string) (*ExpertTemplate, error) {
	raw, err := c.apiGet(fmt.Sprintf("/api/experts/%s", id))
	if err != nil {
		return nil, err
	}
	var exp ExpertTemplate
	if err := json.Unmarshal(raw, &exp); err != nil {
		return nil, err
	}
	return &exp, nil
}

func (c *OctopClient) CreateAgentFromExpert(expertID string, req CreateAgentFromExpertRequest) (*AgentSummary, error) {
	raw, err := c.apiPost(fmt.Sprintf("/api/agents/from-expert/%s", expertID), req)
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	aid, _ := res["agent_id"].(string)
	name, _ := res["name"].(string)
	desc, _ := res["description"].(string)
	model, _ := res["default_model"].(string)
	state, _ := res["state"].(string)
	kind, _ := res["kind"].(string)
	icon, _ := res["icon_name"].(string)
	color, _ := res["color"].(string)

	return &AgentSummary{
		AgentID:      aid,
		Name:         name,
		Description:  desc,
		DefaultModel: model,
		State:        state,
		Kind:         kind,
		Enabled:      true,
		IconName:     icon,
		Color:        color,
	}, nil
}

// --- Teams API ---

func (c *OctopClient) ListTeams() ([]TeamPayload, error) {
	raw, err := c.apiGet("/api/teams")
	if err != nil {
		return nil, err
	}
	var list []TeamPayload
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *OctopClient) GetTeam(teamID string) (*TeamPayload, error) {
	raw, err := c.apiGet(fmt.Sprintf("/api/teams/%s", teamID))
	if err != nil {
		return nil, err
	}
	var team TeamPayload
	if err := json.Unmarshal(raw, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *OctopClient) CreateTeam(req CreateTeamRequest) (*TeamPayload, error) {
	raw, err := c.apiPost("/api/teams", req)
	if err != nil {
		return nil, err
	}
	var team TeamPayload
	if err := json.Unmarshal(raw, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *OctopClient) UpdateTeam(teamID string, req UpdateTeamRequest) (*TeamPayload, error) {
	raw, err := c.apiPatch(fmt.Sprintf("/api/teams/%s", teamID), req)
	if err != nil {
		return nil, err
	}
	var team TeamPayload
	if err := json.Unmarshal(raw, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (c *OctopClient) DeleteTeam(teamID string) error {
	return c.apiDelete(fmt.Sprintf("/api/teams/%s", teamID))
}

func (c *OctopClient) ResolveTeam(target string) (*TeamPayload, error) {
	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}
	ident := strings.TrimSpace(target)
	if ident == "" {
		if len(teams) > 0 {
			return &teams[0], nil
		}
		return nil, fmt.Errorf("no expert teams found")
	}

	// 1. Exact ID
	for _, t := range teams {
		if t.TeamID == ident || t.AgentID == ident {
			return &t, nil
		}
	}

	// 2. Exact Name
	for _, t := range teams {
		if strings.EqualFold(t.Name, ident) {
			return &t, nil
		}
	}

	// 3. Substring Name
	for _, t := range teams {
		if strings.Contains(strings.ToLower(t.Name), strings.ToLower(ident)) {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("expert team %q not found", target)
}
