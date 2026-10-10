package client

type AgentSummary struct {
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultModel string `json:"default_model,omitempty"`
	Enabled      bool   `json:"enabled"`
	IsShared     bool   `json:"is_shared,omitempty"`
	IconName     string `json:"icon_name,omitempty"`
}

type ThreadSummary struct {
	ThreadID  string `json:"thread_id"`
	Title     string `json:"title,omitempty"`
	Preview   string `json:"preview,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type SkillSummary struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
	Kind          string `json:"kind"` // "builtin" | "workspace"
	DisplayNameZH string `json:"display_name_zh,omitempty"`
	DisplayNameEN string `json:"display_name_en,omitempty"`
}

type SkillFilePart struct {
	Path          string `json:"path"`
	ContentBase64 string `json:"content_base64"`
}

type CreateSkillRequest struct {
	Name      string          `json:"name"`
	Content   string          `json:"content,omitempty"`
	Files     []SkillFilePart `json:"files,omitempty"`
	Overwrite bool            `json:"overwrite"`
}

type TurnResult struct {
	Status          string   `json:"status"`
	AgentID         string   `json:"agent_id"`
	AgentName       string   `json:"agent_name"`
	ThreadID        string   `json:"thread_id"`
	Response        string   `json:"response"`
	ToolsCalled     []string `json:"tools_called"`
	DurationSeconds float64  `json:"duration_seconds"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type InboundTurnPayload struct {
	Type             string `json:"type"`
	Text             string `json:"text"`
	ThreadID         string `json:"thread_id,omitempty"`
	ConversationMode string `json:"conversation_mode,omitempty"`
}
