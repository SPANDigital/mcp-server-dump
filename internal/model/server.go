package model

// ServerInfo represents information about an MCP server
type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
	Tools        []Tool       `json:"tools"`
	Resources    []Resource   `json:"resources"`
	Prompts      []Prompt     `json:"prompts"`
	ToolCalls    []ToolCall   `json:"toolCalls,omitempty"`
}

// Capabilities represents the capabilities of an MCP server
type Capabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema any               `json:"inputSchema"`
	Context     map[string]string `json:"context,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MimeType    string            `json:"mimeType"`
	Context     map[string]string `json:"context,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   []any             `json:"arguments"`
	Context     map[string]string `json:"context,omitempty"`
}

// ToolCall represents the result of calling an MCP tool
type ToolCall struct {
	ToolName          string `json:"toolName"`
	Arguments         any    `json:"arguments,omitempty"`
	Content           []any  `json:"content,omitempty"`
	StructuredContent any    `json:"structuredContent,omitempty"`
	Error             string `json:"error,omitempty"`
}
