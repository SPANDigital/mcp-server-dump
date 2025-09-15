package app

import (
	"errors"
	"time"

	"github.com/alecthomas/kong"
)

// CLI represents the command line interface configuration
type CLI struct {
	// Version flag
	Version kong.VersionFlag `kong:"short='v',help='Show version information'"`

	// Output options
	Output string `kong:"short='o',help='Output file for documentation (defaults to stdout)'"`
	Format string `kong:"short='f',default='markdown',enum='markdown,json,html,pdf',help='Output format'"`
	NoTOC  bool   `kong:"help='Disable table of contents in markdown output'"`

	// Frontmatter options
	Frontmatter       bool     `kong:"short='F',help='Include frontmatter in markdown output'"`
	FrontmatterField  []string `kong:"short='M',help='Add custom frontmatter field (format: key:value), can be used multiple times'"`
	FrontmatterFormat string   `kong:"default='yaml',enum='yaml,toml,json',help='Frontmatter format'"`

	// Transport selection
	Transport string `kong:"short='t',default='command',enum='command,sse,streamable',help='Transport type'"`

	// Transport-specific options
	Endpoint string        `kong:"help='HTTP endpoint for SSE/Streamable transports'"`
	Timeout  time.Duration `kong:"default='30s',help='HTTP timeout for SSE/Streamable transports'"`
	Headers  []string      `kong:"short='H',help='HTTP headers for SSE/Streamable transports (format: Key:Value)'"`

	// Context configuration
	ContextFile   []string `kong:"help='Path to context configuration files (YAML/JSON), can be used multiple times'"`
	ServerCommand string   `kong:"help='Server command for explicit command transport'"`

	// Scanning options
	NoTools     bool `kong:"help='Skip scanning tools from the MCP server'"`
	NoResources bool `kong:"help='Skip scanning resources from the MCP server'"`
	NoPrompts   bool `kong:"help='Skip scanning prompts from the MCP server'"`

	// Legacy command format (backward compatibility)
	Args []string `kong:"arg,optional,help='Command and arguments (legacy format for backward compatibility)'"`
}

// ValidateScanOptions validates that at least one scan type is enabled
func (cli *CLI) ValidateScanOptions() error {
	if cli.NoTools && cli.NoResources && cli.NoPrompts {
		return errors.New(ErrAllScanTypesDisabled)
	}
	return nil
}
