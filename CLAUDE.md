# Claude Code Configuration

This file contains configuration and commands for Claude Code to assist with this project.

## Project Overview

mcp-server-dump is a Go-based command-line tool for extracting documentation from MCP (Model Context Protocol) servers. It connects to MCP servers via multiple transports (STDIO/command, SSE, and streamable HTTP) and dumps their capabilities, tools, resources, and prompts to Markdown, JSON, HTML, or PDF format. The tool includes frontmatter support for static site generator integration.

## Development Commands

### Build Commands
```bash
# Build from new entry point
go build -o mcp-server-dump ./cmd/mcp-server-dump

# Or use make-style command for convenience
go build -o mcp-server-dump ./cmd/mcp-server-dump
```

### Test Commands
```bash
go test ./...
```

### Lint Commands
```bash
go fmt ./...
go vet ./...
```

### Dependencies
```bash
go mod tidy
go mod download
```

### Linting
```bash
# Install golangci-lint v2.4.0 (required for Go 1.25 support)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0

# Run linter (may need to use $GOPATH/bin/golangci-lint if not in PATH)
golangci-lint run
```

## Key Files

### Main Entry Point
- `cmd/mcp-server-dump/main.go` - Minimal main function that delegates to internal packages

### Internal Packages (Private API)
- `internal/app/` - Application logic and CLI configuration
  - `cli.go` - Command line interface definition
  - `runner.go` - Main application logic and MCP client implementation
  - `version.go` - Version handling with build info detection
  - `templates.go` - Embedded template filesystem
  - `templates/` - Go template files for markdown output formatting
- `internal/transport/` - MCP transport implementations
  - `factory.go` - Transport factory with configuration
  - `header.go` - HTTP header middleware
  - `content_fix.go` - Content-type normalization for streamable transport
- `internal/formatter/` - Output formatters
  - `markdown.go` - Markdown formatting with templates
  - `json.go` - JSON output formatting
  - `html.go` - HTML output (converts markdown via Goldmark)
  - `pdf.go` - PDF generation using go-pdf/fpdf
  - `frontmatter.go` - YAML/TOML/JSON frontmatter generation
  - `utils.go` - Common formatting utilities
- `internal/model/` - Data structures
  - `server.go` - ServerInfo, Tool, Resource, Prompt, Capabilities structs

### Configuration Files
- `go.mod` - Go module dependencies
- `.gitignore` - Git ignore patterns for Go projects  
- `.goreleaser.yaml` - GoReleaser configuration for multi-platform builds
- `README.md` - Project documentation

## Dependencies

### Production Dependencies
- `github.com/modelcontextprotocol/go-sdk` - Official MCP Go SDK for client/server communication
- `github.com/alecthomas/kong` - Command line argument parsing library
- `github.com/yuin/goldmark` - Markdown to HTML converter with GitHub Flavored Markdown support
- `github.com/go-pdf/fpdf` - Pure Go PDF generation library with bookmark support
- `gopkg.in/yaml.v2` - YAML parsing and generation for frontmatter support

### Development Tools
- Go 1.25.0+ - Required Go version
- Standard Go toolchain (go fmt, go vet, go test)

## Usage Examples

### Basic Usage
```bash
# Connect to filesystem server (command transport)
mcp-server-dump npx @modelcontextprotocol/server-filesystem /Users/username/Documents

# Connect to custom Node.js server
mcp-server-dump node server.js --port 3000

# Connect via SSE transport with headers
mcp-server-dump -t sse --endpoint "http://localhost:3001/sse" -H "Authorization:Bearer token"

# Connect via streamable transport
mcp-server-dump -t streamable --endpoint "http://localhost:3001/stream"

# Disable table of contents in markdown output
mcp-server-dump --no-toc node server.js

# Generate HTML output
mcp-server-dump -f html node server.js

# Generate PDF output (requires output file)
mcp-server-dump -f pdf -o server-docs.pdf node server.js

# Output to JSON file
mcp-server-dump -f json -o output.json python mcp_server.py
```

### Development Testing
```bash
# Build and test with example server
go build -o mcp-server-dump ./cmd/mcp-server-dump
mcp-server-dump echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}''
```

## Architecture

The tool follows a modern layered architecture with clear separation of concerns:

### Package Structure
1. **Entry Point (`cmd/`)** - Minimal main function that delegates to application logic
2. **Application Layer (`internal/app/`)** - CLI parsing, version handling, and orchestration
3. **Transport Layer (`internal/transport/`)** - MCP transport implementations with factory pattern
4. **Formatting Layer (`internal/formatter/`)** - Output format implementations 
5. **Model Layer (`internal/model/`)** - Data structures and type definitions

### Processing Flow
1. **CLI Parsing** - Kong library handles command line argument parsing (internal/app/cli.go)
2. **Transport Creation** - Factory pattern creates appropriate transport with middleware:
   - Command transport: executes MCP server as subprocess with STDIO communication
   - SSE transport: connects to HTTP Server-Sent Events endpoints  
   - Streamable transport: connects to HTTP streamable endpoints with content-type fixing
   - HTTP Headers Support: Custom headers via HeaderRoundTripper middleware
3. **MCP Communication** - JSON-RPC over the selected transport using official MCP Go SDK
4. **Data Extraction** - Session-based communication to list tools, resources, prompts
5. **Output Formatting** - Pluggable formatters convert server data to various formats:
   - Markdown: Go text/template with embedded template files and TOC support
   - HTML: Generated from Markdown using Goldmark with GitHub Flavored Markdown extensions
   - JSON: Direct marshaling of data structures
   - PDF: Generated using go-pdf/fpdf with structured layout
   - Frontmatter: YAML/TOML/JSON metadata for static site generators

## Code Structure

```
cmd/mcp-server-dump/
└── main.go - Minimal main function that calls app.Run()

internal/
├── app/
│   ├── cli.go - Kong CLI configuration struct
│   ├── runner.go - Main application logic and MCP client orchestration
│   ├── version.go - Version detection using build info and VCS data
│   ├── templates.go - Embedded template filesystem
│   └── templates/
│       ├── base.md.tmpl - Main template with TOC structure
│       ├── capabilities.md.tmpl - Capabilities section with emoji indicators
│       ├── tools.md.tmpl - Tools listing with anchored headings  
│       ├── resources.md.tmpl - Resources section
│       └── prompts.md.tmpl - Prompts section
├── transport/
│   ├── factory.go - Transport factory with config-driven creation
│   ├── header.go - HeaderRoundTripper for custom HTTP headers
│   └── content_fix.go - Content-type normalization middleware
├── formatter/
│   ├── markdown.go - Template-based markdown generation
│   ├── json.go - JSON marshaling formatter
│   ├── html.go - Goldmark-based HTML converter (markdown → HTML)
│   ├── pdf.go - PDF generation using go-pdf/fpdf
│   ├── frontmatter.go - YAML/TOML/JSON frontmatter generation
│   └── utils.go - Shared formatting utilities (anchors, JSON indent)
└── model/
    └── server.go - Data structures (ServerInfo, Tool, Resource, Prompt, Capabilities)
```

### Key Design Patterns
- **Factory Pattern**: Transport creation based on configuration
- **Middleware Pattern**: HTTP transport chain (base → headers → content-fix)
- **Template Pattern**: Pluggable formatters with common interface
- **Embedded Assets**: Templates embedded in binary via go:embed
- **Layered Architecture**: Clear separation between CLI, business logic, and I/O

## Troubleshooting

### Common Issues

1. **Build Errors**: Run `go mod tidy` to fix dependency issues
2. **Connection Failures**: 
   - For command transport: Ensure target MCP server is executable and supports STDIO
   - For HTTP transports: Verify endpoint URL is correct and server is running
3. **Permission Errors**: Check file permissions for output directory
4. **HTTP Header Issues**: Ensure headers are in correct Key:Value format

### Debug Commands
```bash
# Check if binary works
mcp-server-dump -h

# Test with verbose output
mcp-server-dump -v node server.js 2>&1 | tee debug.log
```

## Contributing

When making changes:

1. Update README.md with new features or usage changes
2. Run `go fmt`, `go vet`, and `golangci-lint run` before committing
3. Test with multiple MCP server implementations
4. Update this CLAUDE.md file for significant architectural changes
5. When adding new packages, follow the internal/ structure for private code
6. Use the factory pattern for extensible components (transports, formatters)
7. Keep the cmd/ main.go minimal - business logic belongs in internal/app/
- This project should not have a Dockerfile, I will use ko via goreleaser to build container images
- Wherever possible use Context7, go doc, and github to source the latest documentation
- GHCR container repository owner is spandigital, not richardwooding
- Commands installed via "go install" are located in $GOPATH/bin if not in PATH
- Alway use any instead of interface{}. In modern go, any is an alias for interface{}
- The correct way to install goreleaser is "go install github.com/goreleaser/goreleaser/v2@latest"