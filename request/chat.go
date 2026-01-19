package request

type ChatRequest struct {
	SessionID                    string      `json:"session_id"`
	Query                        string      `json:"query"`
	AgentConfig                  AgentConfig `json:"agent_config"`
	UploadedFiles                []string    `json:"uploaded_files"`
	EnableKnowledgeBaseRetrieval bool        `json:"enable_knowledge_base_retrieval"`
}

type AgentConfig struct {
	Model         string   `json:"model"`
	MaxIterations int      `json:"max_iterations"`
	Tools         []string `json:"tools"`
}
