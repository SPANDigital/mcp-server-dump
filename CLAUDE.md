# Claude Code Configuration

This file contains configuration and commands for Claude Code to assist with this project.

## Project Overview

mcp-server-dump is a Go-based command-line tool for extracting documentation from MCP (Model Context Protocol) servers. It connects to MCP servers via multiple transports (STDIO/command, SSE, and streamable HTTP) and dumps their capabilities, tools, resources, and prompts to Markdown, JSON, HTML, or PDF format.

## Development Commands

### Build Commands
```bash
go build -o mcp-server-dump
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

- `main.go` - Main application logic and MCP client implementation
- `go.mod` - Go module dependencies
- `README.md` - Project documentation
- `.gitignore` - Git ignore patterns for Go projects
- `templates/` - Go template files for markdown output formatting
  - `base.md.tmpl` - Main document structure with Table of Contents
  - `capabilities.md.tmpl` - Server capabilities section
  - `tools.md.tmpl` - Tools listing template
  - `resources.md.tmpl` - Resources section template
  - `prompts.md.tmpl` - Prompts section template

## Dependencies

### Production Dependencies
- `github.com/modelcontextprotocol/go-sdk` - Official MCP Go SDK for client/server communication
- `github.com/alecthomas/kong` - Command line argument parsing library
- `github.com/yuin/goldmark` - Markdown to HTML converter with GitHub Flavored Markdown support
- `github.com/johnfercher/maroto/v2` - Pure Go PDF generation library

### Development Tools
- Go 1.25.0+ - Required Go version
- Standard Go toolchain (go fmt, go vet, go test)

## Usage Examples

### Basic Usage
```bash
# Connect to filesystem server (command transport)
./mcp-server-dump npx @modelcontextprotocol/server-filesystem /Users/username/Documents

# Connect to custom Node.js server
./mcp-server-dump node server.js --port 3000

# Connect via SSE transport with headers
./mcp-server-dump -t sse --endpoint "http://localhost:3001/sse" -H "Authorization:Bearer token"

# Connect via streamable transport
./mcp-server-dump -t streamable --endpoint "http://localhost:3001/stream"

# Disable table of contents in markdown output
./mcp-server-dump --no-toc node server.js

# Generate HTML output
./mcp-server-dump -f html node server.js

# Generate PDF output (requires output file)
./mcp-server-dump -f pdf -o server-docs.pdf node server.js

# Output to JSON file
./mcp-server-dump -f json -o output.json python mcp_server.py
```

### Development Testing
```bash
# Build and test with example server
go build -o mcp-server-dump
./mcp-server-dump echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}'
```

## Architecture

The tool uses the following architecture:

1. **CLI Parsing** - Kong library handles command line argument parsing
2. **Transport Selection** - Supports multiple transport types:
   - Command transport: executes MCP server as subprocess with STDIO communication
   - SSE transport: connects to HTTP Server-Sent Events endpoints
   - Streamable transport: connects to HTTP streamable endpoints
3. **HTTP Headers Support** - Custom headers for SSE/streamable transports via HeaderRoundTripper
4. **MCP Communication** - JSON-RPC over the selected transport
5. **Data Extraction** - Lists tools, resources, prompts via MCP protocol methods
6. **Output Formatting** - Converts server data to Markdown, JSON, or HTML format
   - Markdown uses Go text/template with embedded template files
   - HTML generated from Markdown using Goldmark with GitHub Flavored Markdown extensions
   - Templates support Table of Contents with anchor links
   - Custom template functions for formatting (anchor generation, JSON indentation)

## Code Structure

```
main.go
├── CLI struct - Kong configuration for command line interface (includes Headers field)
├── ServerInfo struct - Internal representation of MCP server data
├── HeaderRoundTripper - HTTP RoundTripper for adding custom headers
├── parseHeaders() - Parses header strings in Key:Value format
├── createTransport() - Creates appropriate transport (command/SSE/streamable) with headers
├── run() - Main application logic
│   ├── Transport creation with header support
│   ├── MCP client connection
│   ├── Server capability introspection
│   └── Output formatting
├── formatMarkdown() - Template-based markdown formatter
├── formatHTML() - Goldmark-based HTML formatter (converts markdown to HTML)
├── anchorName() - Convert strings to URL-safe anchors
└── jsonIndent() - Format JSON with indentation

templates/
├── base.md.tmpl - Main template with TOC structure
├── capabilities.md.tmpl - Capabilities section with emoji indicators
├── tools.md.tmpl - Tools listing with anchored headings
├── resources.md.tmpl - Resources section
└── prompts.md.tmpl - Prompts section
```

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
./mcp-server-dump -h

# Test with verbose output
./mcp-server-dump -v node server.js 2>&1 | tee debug.log
```

## Contributing

When making changes:

1. Update README.md with new features or usage changes
2. Run `go fmt` and `go vet` before committing
3. Test with multiple MCP server implementations
4. Update this CLAUDE.md file for significant architectural changes
- This project should not have a Dockerfile, I will use ko via goreleaser to build container images
- Wherever possible use Context7, go doc, and github to source the latest documentation
- GHCR container repository owner is spandigital, not richardwooding
- Commands installed via "go install" are located in $GOPATH/bin if not in PATH
- Alway use any instead of interface{}. In modern go, any is an alias for interface{}