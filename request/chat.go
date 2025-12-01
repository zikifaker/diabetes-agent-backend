package request

type ChatRequest struct {
	SessionID   string      `json:"session_id"`
	Query       string      `json:"query"`
	AgentConfig AgentConfig `json:"agent_config"`
}

type AgentConfig struct {
	Model         string   `json:"model"`
	MaxIterations int      `json:"max_iterations"`
	Tools         []string `json:"tools"`
}
