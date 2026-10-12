package client

type AgentSummary struct {
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultModel string `json:"default_model,omitempty"`
	State        string `json:"state,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Enabled      bool   `json:"enabled"`
	IsShared     bool   `json:"is_shared,omitempty"`
	IconName     string `json:"icon_name,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	Color        string `json:"color,omitempty"`
}

type ChannelSummary struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type AgentRuntimeStatus struct {
	State     string           `json:"state"`
	LastError *string          `json:"last_error,omitempty"`
	Channels  []ChannelSummary `json:"channels,omitempty"`
	CronJobs  []interface{}    `json:"cron_jobs,omitempty"`
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

type LocalizedText struct {
	ZH string `json:"zh,omitempty"`
	EN string `json:"en,omitempty"`
}

type ExpertTemplate struct {
	ID             string        `json:"id"`
	Label          LocalizedText `json:"label"`
	Description    LocalizedText `json:"description"`
	WelcomeMessage LocalizedText `json:"welcome_message,omitempty"`
	IconName       string        `json:"icon_name,omitempty"`
	IconURL        string        `json:"icon_url,omitempty"`
	Color          string        `json:"color,omitempty"`
}

type CreateAgentFromExpertRequest struct {
	Name             string   `json:"name,omitempty"`
	Description      string   `json:"description,omitempty"`
	DefaultModel     string   `json:"default_model,omitempty"`
	Color            string   `json:"color,omitempty"`
	WelcomeMessage   string   `json:"welcome_message,omitempty"`
	ConversationMode string   `json:"conversation_mode,omitempty"`
	Providers        []string `json:"providers,omitempty"`
}

type TeamMember struct {
	AgentID  string `json:"agent_id"`
	Name     string `json:"name"`
	Color    string `json:"color,omitempty"`
	IconName string `json:"icon_name,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
	State    string `json:"state,omitempty"`
	IsShared bool   `json:"is_shared,omitempty"`
	UserID   *int   `json:"user_id,omitempty"`
}

type TeamPayload struct {
	TeamID         string       `json:"team_id"`
	AgentID        string       `json:"agent_id"`
	Name           string       `json:"name"`
	Description    string       `json:"description,omitempty"`
	DefaultModel   string       `json:"default_model,omitempty"`
	Color          string       `json:"color,omitempty"`
	IconName       string       `json:"icon_name,omitempty"`
	IconURL        string       `json:"icon_url,omitempty"`
	WelcomeMessage string       `json:"welcome_message,omitempty"`
	State          string       `json:"state,omitempty"`
	Kind           string       `json:"kind,omitempty"`
	MemberIDs      []string     `json:"member_ids"`
	Members        []TeamMember `json:"members,omitempty"`
}

type CreateTeamRequest struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	DefaultModel   string                 `json:"default_model,omitempty"`
	Color          string                 `json:"color,omitempty"`
	IconName       string                 `json:"icon_name,omitempty"`
	IconURL        string                 `json:"icon_url,omitempty"`
	WelcomeMessage string                 `json:"welcome_message,omitempty"`
	MemberIDs      []string               `json:"member_ids,omitempty"`
	Config         map[string]interface{} `json:"config,omitempty"`
}

type UpdateTeamRequest struct {
	Name           *string   `json:"name,omitempty"`
	Description    *string   `json:"description,omitempty"`
	DefaultModel   *string   `json:"default_model,omitempty"`
	Color          *string   `json:"color,omitempty"`
	IconName       *string   `json:"icon_name,omitempty"`
	WelcomeMessage *string   `json:"welcome_message,omitempty"`
	MemberIDs      *[]string `json:"member_ids,omitempty"`
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
