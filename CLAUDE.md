# Claude Code Configuration

This file contains configuration and commands for Claude Code to assist with this project.

## Project Overview

mcp-server-dump is a Go-based command-line tool for extracting documentation from MCP (Model Context Protocol) servers. It connects to MCP servers via STDIO and dumps their capabilities, tools, resources, and prompts to Markdown or JSON format.

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

### Development Tools
- Go 1.25.0+ - Required Go version
- Standard Go toolchain (go fmt, go vet, go test)

## Usage Examples

### Basic Usage
```bash
# Connect to filesystem server
./mcp-server-dump npx @modelcontextprotocol/server-filesystem /Users/username/Documents

# Connect to custom Node.js server
./mcp-server-dump node server.js --port 3000

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
2. **MCP Connection** - CommandTransport executes the target MCP server as subprocess
3. **STDIO Communication** - JSON-RPC over stdin/stdout with the MCP server
4. **Data Extraction** - Lists tools, resources, prompts via MCP protocol methods
5. **Output Formatting** - Converts server data to Markdown or JSON format
   - Markdown uses Go text/template with embedded template files
   - Templates support Table of Contents with anchor links
   - Custom template functions for formatting (anchor generation, JSON indentation)

## Code Structure

```
main.go
├── CLI struct - Kong configuration for command line interface
├── ServerInfo struct - Internal representation of MCP server data
├── run() - Main application logic
│   ├── CommandTransport creation
│   ├── MCP client connection
│   ├── Server capability introspection
│   └── Output formatting
├── formatMarkdown() - Template-based markdown formatter
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
2. **Connection Failures**: Ensure target MCP server is executable and supports STDIO
3. **Permission Errors**: Check file permissions for output directory

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