package model

// ServerInfo represents information about an MCP server
type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Capabilities Capabilities `json:"capabilities"`
	Tools        []Tool       `json:"tools"`
	Resources    []Resource   `json:"resources"`
	Prompts      []Prompt     `json:"prompts"`
}

// Capabilities represents the capabilities of an MCP server
type Capabilities struct {
	Tools     bool `json:"tools"`
	Resources bool `json:"resources"`
	Prompts   bool `json:"prompts"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   []any  `json:"arguments"`
}
