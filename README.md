# mcp-server-dump

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.25.0%2B-blue)](https://go.dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/spandigital/mcp-server-dump)](https://goreportcard.com/report/github.com/spandigital/mcp-server-dump)
[![Release](https://img.shields.io/github/v/release/spandigital/mcp-server-dump)](https://github.com/spandigital/mcp-server-dump/releases)

A command-line tool to extract and document MCP (Model Context Protocol) server capabilities, tools, resources, and prompts in various formats.

## Features

- **Multiple Transport Support**: Connect to MCP servers via various transports:
  - STDIO/Command transport (subprocess execution)  
  - Streamable HTTP transport
  - Server-Sent Events (SSE) over HTTP *(deprecated)*
- Extract server information, capabilities, tools, resources, and prompts
- Output documentation in Markdown, JSON, HTML, or PDF format *(PDF format is WIP)*
- **Enhanced Markdown output with clickable Table of Contents**
- **Frontmatter support** for static site generator integration (Hugo, Jekyll, etc.)
- **External Go templates for customizable documentation**
- **Backward compatible** with existing command-line usage
- Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Clean CLI interface powered by [Kong](https://github.com/alecthomas/kong)

## Installation

### Using Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap spandigital/homebrew-tap

# Install the tool
brew install mcp-server-dump
```

Or install directly in one command:
```bash
brew install spandigital/homebrew-tap/mcp-server-dump
```

### Using Scoop (Windows)

```powershell
# Add the bucket
scoop bucket add spandigital https://github.com/spandigital/scoop-bucket

# Install the tool
scoop install mcp-server-dump
```

### Using go install

```bash
go install github.com/spandigital/mcp-server-dump/cmd/mcp-server-dump@latest
```

The binary will be installed to `$GOPATH/bin/mcp-server-dump` (or `$(go env GOPATH)/bin/mcp-server-dump`). Make sure your Go bin directory is in your PATH.

### From Source

```bash
git clone https://github.com/spandigital/mcp-server-dump.git
cd mcp-server-dump
go build -o mcp-server-dump ./cmd/mcp-server-dump
```

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/spandigital/mcp-server-dump/releases) for your platform.

### Requirements

- Go 1.25.0 or later
- Access to MCP servers (Node.js, Python, etc.)

## Usage

### Basic Usage

```bash
# Connect to a Node.js MCP server (default command transport)
mcp-server-dump node server.js

# Connect to a Python MCP server with arguments
mcp-server-dump python server.py --config config.json

# Connect to an NPX package
mcp-server-dump npx @modelcontextprotocol/server-filesystem /path/to/directory

# Connect to a UVX package (Python equivalent of npx)
mcp-server-dump uvx mcp-server-sqlite --db-path /path/to/database.db

# Run a Go MCP server directly
mcp-server-dump go run github.com/example/mcp-server@latest --example-argument=something

# Connect to a streamable HTTP transport server
mcp-server-dump --transport=streamable --endpoint="http://localhost:3001/stream"

# Connect to a streamable HTTP transport server (alternative endpoint)
mcp-server-dump --transport=streamable --endpoint="http://localhost:8080/mcp"
```

### Transport Options

```bash
# Command transport (default) - runs server as subprocess
mcp-server-dump --transport=command node server.js
mcp-server-dump --transport=command --server-command="python server.py --arg value"

# Streamable transport - connects to HTTP streamable endpoint
mcp-server-dump --transport=streamable --endpoint="http://localhost:3001/stream"

# Configure HTTP timeout for web transports
mcp-server-dump --transport=streamable --endpoint="http://localhost:3001/stream" --timeout=60s

# Add custom HTTP headers for authentication or other purposes
mcp-server-dump --transport=streamable --endpoint="http://localhost:3001/stream" \
  -H "Authorization:Bearer your-token-here" \
  -H "X-API-Key:your-api-key"

# Disable table of contents in markdown output
mcp-server-dump --no-toc node server.js

# Generate HTML output from markdown
mcp-server-dump -f html node server.js

# HTML output without table of contents
mcp-server-dump -f html --no-toc node server.js

# Generate PDF output (requires output file)
mcp-server-dump -f pdf -o server-docs.pdf node server.js

# PDF output without table of contents
mcp-server-dump -f pdf --no-toc -o server-docs.pdf node server.js
```

### Output Options

```bash
# Output to file (Markdown by default)
mcp-server-dump -o server-docs.md node server.js

# JSON output
mcp-server-dump -f json node server.js

# HTML output
mcp-server-dump -f html node server.js

# PDF output (requires output file)
mcp-server-dump -f pdf -o server-docs.pdf node server.js

# Output to file (any format)
mcp-server-dump -f json -o server-info.json python server.py
mcp-server-dump -f html -o server-docs.html python server.py
mcp-server-dump -f pdf -o server-docs.pdf python server.py
```

### Frontmatter Support

Generate YAML, TOML, or JSON frontmatter in markdown output for integration with static site generators:

```bash
# Basic frontmatter (YAML by default)
mcp-server-dump --frontmatter node server.js

# With custom metadata fields
mcp-server-dump --frontmatter \
  -M "author:joe.bloggs@company.com" \
  -M "status:draft" \
  -M "team:engineering" \
  -M "tags:mcp,documentation,tools" \
  node server.js

# TOML format frontmatter
mcp-server-dump --frontmatter --frontmatter-format=toml \
  -M "author:jane.doe@example.org" \
  -M "reviewed:false" \
  python server.py

# JSON format frontmatter
mcp-server-dump --frontmatter --frontmatter-format=json \
  -M "build_number:42" \
  -M "environment:production" \
  npx @modelcontextprotocol/server-filesystem /path
```

**Auto-generated fields:**
- `title`: Server name + " Documentation"
- `version`: Server version (if available)
- `generated_at`: ISO 8601 timestamp
- `generator`: "mcp-server-dump"
- `capabilities`: Object with tools/resources/prompts booleans
- `tools_count`, `resources_count`, `prompts_count`: Counts of each type

**Custom fields features:**
- Automatic type detection (boolean, integer, float, string)
- Comma-separated values become arrays: `tags:one,two,three`
- Custom fields override auto-generated ones if same key is used

### Command Line Options

```
Usage: mcp-server-dump [<args> ...] [flags]

Arguments:
  [<args> ...]               Command and arguments (legacy format for backward compatibility)

Flags:
  -h, --help                 Show context-sensitive help
  -o, --output=STRING        Output file for documentation (defaults to stdout)
  -f, --format="markdown"    Output format (markdown, json, html, pdf)
      --no-toc               Disable table of contents in markdown output
  -F, --frontmatter          Include frontmatter in markdown output
  -M, --frontmatter-field=FIELD,...
                             Add custom frontmatter field (format: key:value), can be used multiple times
      --frontmatter-format="yaml"
                             Frontmatter format (yaml, toml, json)
  -t, --transport="command"  Transport type (command, sse, streamable)
      --endpoint=STRING      HTTP endpoint for SSE/Streamable transports
      --timeout=30s          HTTP timeout for SSE/Streamable transports
  -H, --headers=HEADERS,...  HTTP headers for SSE/Streamable transports (format: Key:Value)
      --server-command=STRING Server command for explicit command transport
```

## GitHub Action

This tool is also available as a GitHub Action for automated documentation generation in CI/CD pipelines.

### Basic Usage

```yaml
name: Generate MCP Server Documentation
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Generate MCP Server Documentation
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'node server.js'
          format: 'markdown'
          output-file: 'docs/server-documentation.md'
      
      - name: Upload documentation
        uses: actions/upload-artifact@v4
        with:
          name: mcp-server-docs
          path: docs/server-documentation.md
```

### Advanced Usage

```yaml
name: Multi-format Documentation Generation
on:
  release:
    types: [ published ]

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        format: [markdown, html, json, pdf]
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Generate MCP Documentation (${{ matrix.format }})
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'npx @modelcontextprotocol/server-filesystem ./data'
          format: ${{ matrix.format }}
          output-file: 'docs/server.${{ matrix.format == "markdown" && "md" || matrix.format }}'
          frontmatter: 'yaml'
          verbose: 'true'
      
      - name: Upload ${{ matrix.format }} documentation
        uses: actions/upload-artifact@v4
        with:
          name: mcp-docs-${{ matrix.format }}
          path: docs/server.*
```

### HTTP Transport Usage

```yaml
name: Document Remote MCP Server
on:
  workflow_dispatch:
    inputs:
      endpoint:
        description: 'MCP Server Endpoint'
        required: true
        default: 'http://localhost:3001/stream'
      auth_token:
        description: 'Authentication Token'
        required: false

jobs:
  document-server:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      
      - name: Generate Remote Server Documentation
        uses: spandigital/mcp-server-dump@v1
        with:
          transport: 'streamable'
          endpoint: ${{ github.event.inputs.endpoint }}
          headers: 'Authorization:Bearer ${{ github.event.inputs.auth_token }}'
          format: 'html'
          output-file: 'remote-server-docs.html'
```

### Action Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `server-command` | MCP server command to execute | No | - |
| `transport` | Transport type (stdio, sse, streamable) | No | `stdio` |
| `endpoint` | Endpoint URL for sse or streamable transport | No | - |
| `headers` | HTTP headers in Key:Value format (comma-separated) | No | - |
| `format` | Output format (markdown, html, json, pdf) | No | `markdown` |
| `output-file` | Output file path (required for pdf format) | No | - |
| `no-toc` | Disable table of contents in markdown output | No | `false` |
| `frontmatter` | Add frontmatter to output (yaml, toml, json) | No | - |
| `timeout` | Connection timeout in seconds | No | `30` |
| `verbose` | Enable verbose output | No | `false` |

### Action Outputs

| Output | Description |
|--------|-------------|
| `output-file` | Path to the generated output file |
| `server-info` | JSON string containing server capabilities and metadata |

## Use Cases

- **Server Documentation**: Generate comprehensive documentation for MCP servers
- **API Discovery**: Understand available tools and resources in MCP servers
- **Integration Planning**: Analyze server capabilities before integration
- **Reverse Engineering**: Explore and document third-party MCP servers
- **Testing & Validation**: Verify server implementations match specifications
- **CI/CD Integration**: Automated documentation generation in GitHub workflows

## Dependencies

- [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - Official MCP Go SDK
- [github.com/alecthomas/kong](https://github.com/alecthomas/kong) - Command line parser
- [github.com/johnfercher/maroto/v2](https://github.com/johnfercher/maroto/v2) - Pure Go PDF generation library
- [github.com/yuin/goldmark](https://github.com/yuin/goldmark) - Markdown to HTML converter with GitHub Flavored Markdown support

## Template Customization

The markdown output is generated using Go templates located in the `templates/` directory:

- `base.md.tmpl` - Main document structure with Table of Contents
- `capabilities.md.tmpl` - Server capabilities section
- `tools.md.tmpl` - Tools listing with anchored headings
- `resources.md.tmpl` - Resources section
- `prompts.md.tmpl` - Prompts section

You can customize these templates to adjust the output format to your needs. The templates use Go's `text/template` package with custom functions:

- `anchor` - Converts strings to URL-safe anchor names
- `json` - Formats objects as indented JSON

## Deprecated Features

The following features are deprecated and included only for backward compatibility:

### SSE Transport (Server-Sent Events)

⚠️ **Deprecated for new implementations**: While SSE transport continues to work and is supported for backward compatibility, new MCP servers should not be created using SSE. The streamable transport is preferred for new implementations.

```bash
# SSE transport - connects to HTTP Server-Sent Events endpoint
mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse"

# Configure HTTP timeout for SSE transport
mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse" --timeout=60s

# Add custom HTTP headers for SSE transport
mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse" \
  -H "Authorization:Bearer your-token-here" \
  -H "X-API-Key:your-api-key"
```

**Note**: SSE transport remains fully functional for existing servers. Streamable transport is recommended for new server implementations, but is not a drop-in replacement and may require server-side changes.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on how to submit pull requests, report issues, and contribute to the project.

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
