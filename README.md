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
- **Tool Calling**: Call MCP tools directly and include results in documentation
  - Call specific tools by name with custom arguments
  - Call all available tools for comprehensive testing
  - Results automatically integrated into all output formats
- **Selective Scanning**: Skip specific capability types (tools, resources, prompts) for performance optimization
- Output documentation in Markdown, JSON, HTML, PDF, or **Hugo** format
- **Hugo format**: Generate a complete Hugo documentation site structure with hierarchical content organization
- **Enhanced Markdown output with clickable Table of Contents**
- **Rich structured context support** via external YAML/JSON configuration files
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

### Linux Packages

Pre-built Linux packages (.deb, .rpm, .apk) are available on the [releases page](https://github.com/spandigital/mcp-server-dump/releases).

```bash
# Example: Download and install a .deb package
wget https://github.com/spandigital/mcp-server-dump/releases/latest/download/mcp-server-dump_1.20.1_amd64.deb
sudo dpkg -i mcp-server-dump_1.20.1_amd64.deb

# Example: Download and install an .rpm package
wget https://github.com/spandigital/mcp-server-dump/releases/latest/download/mcp-server-dump-1.20.1-1.x86_64.rpm
sudo rpm -i mcp-server-dump-1.20.1-1.x86_64.rpm
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

# Skip scanning specific types (performance optimization)
mcp-server-dump --no-tools node server.js     # Skip tools
mcp-server-dump --no-resources node server.js # Skip resources
mcp-server-dump --no-prompts node server.js   # Skip prompts

# Combine scan flags (at least one type must remain enabled)
mcp-server-dump --no-tools --no-resources node server.js  # Only prompts
```

### Tool Calling

```bash
# Call a specific tool by name
mcp-server-dump --call-tool="get_weather" node server.js

# Call a tool with arguments (JSON format)
mcp-server-dump --call-tool="get_weather" --tool-args='{"location":"London"}' node server.js

# Call multiple specific tools
mcp-server-dump --call-tool="get_weather" --call-tool="get_forecast" node server.js

# Call all available tools (useful for testing/documentation)
mcp-server-dump --call-all-tools node server.js

# Call all tools with specific arguments
mcp-server-dump --call-all-tools --tool-args='{"test":"true"}' node server.js

# Combine tool calling with other options
mcp-server-dump --call-tool="search" --tool-args='{"query":"example"}' -f html -o docs.html node server.js

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

# Hugo documentation site (requires output directory)
mcp-server-dump -f hugo -o hugo-docs node server.js

# Hugo with custom configuration
mcp-server-dump -f hugo -o hugo-docs \
  --hugo-base-url="https://docs.example.com" \
  --hugo-language-code="en-us" \
  --hugo-enterprise-key="my-custom-key" \
  --hugo-author-strict \
  node server.js

# Output to file (any format)
mcp-server-dump -f json -o server-info.json python server.py
mcp-server-dump -f html -o server-docs.html python server.py
mcp-server-dump -f pdf -o server-docs.pdf python server.py
```

### Hugo Documentation Site

The Hugo format generates a complete Hugo site with modern Hugo modules configuration and [Presidium](https://github.com/SPANDigital/presidium-layouts-base) layouts:

```bash
# Generate Hugo documentation site with Presidium layouts
mcp-server-dump -f hugo -o my-docs node server.js

# Directory structure created:
my-docs/
├── hugo.yml                # Hugo configuration with Presidium modules
└── content/
    ├── _index.md           # Root index with server info
    ├── tools/
    │   ├── _index.md      # Tools section index
    │   ├── tool1.md       # Individual tool page
    │   └── tool2.md
    ├── resources/
    │   ├── _index.md      # Resources section index
    │   ├── resource1.md   # Individual resource page
    │   └── resource2.md
    └── prompts/
        ├── _index.md      # Prompts section index
        ├── prompt1.md     # Individual prompt page
        └── prompt2.md

# Build with Hugo (requires Go and Hugo v0.133.0+)
cd my-docs
hugo mod init example.com/my-docs    # Initialize Hugo module
hugo mod get github.com/spandigital/presidium-styling-base
hugo mod get github.com/spandigital/presidium-layouts-base
hugo serve                           # Serve locally at http://localhost:1313
```

**Hugo format features:**
- **Modern Hugo modules**: Uses Hugo modules instead of git submodules for theme management
- **Presidium layouts**: Pre-configured with Presidium documentation layouts optimized for technical documentation
- **Complete Hugo configuration**: Generated `hugo.yml` with Presidium-specific settings (MenuIndex, SearchMap outputs)
- **Hierarchical content organization**: `_index.md` files for sections with proper navigation structure
- **Individual pages**: Separate markdown files for each tool, resource, and prompt
- **SEO-friendly URLs**: Slug-safe filenames and proper Hugo frontmatter
- **Enterprise-ready**: Designed for professional documentation with frontmatter validation
- **Mobile responsive**: Presidium layouts provide excellent mobile experience
- **Search integration**: Built-in search capabilities via SearchMap output format

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

### Rich Structured Context

Add detailed documentation context to tools, resources, and prompts through external YAML/JSON configuration files:

```bash
# Single context file
mcp-server-dump --context-file context.yaml node server.js

# Multiple context files (merged in order)
mcp-server-dump --context-file base.yaml --context-file overrides.json node server.js

# Works with all output formats
mcp-server-dump --context-file context.yaml -f pdf -o docs.pdf node server.js
mcp-server-dump --context-file context.yaml -f html -o docs.html node server.js
```

**Context configuration format:**

```yaml
contexts:
  tools:
    tool_name:
      usage: "Simple usage instruction"
      security: "Security considerations and restrictions"
      examples: |
        Complex multi-line examples with formatting:

        ```json
        {
          "example": "value",
          "options": ["one", "two"]
        }
        ```

        **Additional notes:**
        - Supports markdown formatting
        - Code blocks, lists, tables
        - Multiple paragraphs

  resources:
    "file://*":          # Supports URI patterns with wildcards
      access: "Read-only file system access"
      limitations: |
        **Access restrictions:**
        - Limited to allowed directories
        - Maximum file size: 10MB
        - No write permissions

    "memory://*":
      persistence: "Session-only memory resources"

  prompts:
    prompt_name:
      methodology: |
        **Step-by-step process:**
        1. Input validation and sanitization
        2. Context analysis and processing
        3. Response generation
        4. Quality assurance checks
      output: "Structured JSON response with analysis results"
```

**JSON configuration example:**

```json
{
  "contexts": {
    "tools": {
      "read_file": {
        "usage": "Use this tool to read the contents of files from the filesystem",
        "security": "Only accessible within allowed directories",
        "examples": "```json\\n{\\n  \"path\": \"/home/user/document.txt\"\\n}\\n```",
        "limitations": "- Maximum file size: 1MB\\n- Text files only\\n- Read-only access"
      }
    },
    "resources": {
      "file://*": {
        "description": "Local file system resources",
        "access": "Read-only access to allowed directories"
      },
      "memory://*": {
        "description": "In-memory resources and state",
        "persistence": "Not persisted between sessions"
      }
    },
    "prompts": {
      "analyze_code": {
        "purpose": "Analyze code for security vulnerabilities and best practices",
        "output": "Structured analysis report with severity ratings",
        "parameters": "Required: language, code\\nOptional: focus, severity_filter"
      }
    }
  }
}
```

**Key features:**
- **Multiple formats**: YAML and JSON configuration files supported
- **Rich content**: Multi-line values with full markdown support (code blocks, lists, tables)
- **Smart rendering**: Single-line values as bullet points, multi-line as formatted blocks
- **Pattern matching**: Resources support URI pattern matching with wildcards
- **All output formats**: Context appears in Markdown, HTML, JSON, and PDF
- **InputSchema first**: Context always appears after InputSchema in output
- **Fully optional**: No breaking changes, completely backward compatible

### Selective Scanning Control

Control which types of MCP server capabilities are scanned and included in the documentation for performance optimization and focused documentation:

```bash
# Skip tools entirely (useful for resource-only servers)
mcp-server-dump --no-tools npx @modelcontextprotocol/server-filesystem /docs

# Skip resources entirely (useful for tool-only servers)
mcp-server-dump --no-resources node tool-server.js

# Skip prompts entirely (useful for tool/resource servers)
mcp-server-dump --no-prompts python server.py

# Combine multiple flags (at least one type must remain enabled)
mcp-server-dump --no-tools --no-resources python prompt-server.py  # Only prompts
mcp-server-dump --no-resources --no-prompts node tool-server.js    # Only tools

# Works with all output formats and options
mcp-server-dump --no-tools -f json -o output.json node server.js
mcp-server-dump --no-prompts -f pdf -o docs.pdf python server.py
```

**Key benefits:**
- **Performance**: Skip unnecessary scanning for faster execution
- **Focused documentation**: Generate documentation only for relevant capabilities
- **Bandwidth optimization**: Reduce data transfer for network transports
- **Error isolation**: Continue documentation generation even if specific capability types fail

**Validation**: At least one scan type must remain enabled. The tool will error if all scan types are disabled (`--no-tools --no-resources --no-prompts`).

### Command Line Options

```
Usage: mcp-server-dump [<args> ...] [flags]

Arguments:
  [<args> ...]               Command and arguments (legacy format for backward compatibility)

Flags:
  -h, --help                 Show context-sensitive help
  -o, --output=STRING        Output file for documentation (defaults to stdout, required for hugo format as directory)
  -f, --format="markdown"    Output format (markdown, json, html, pdf, hugo)
      --no-toc               Disable table of contents in markdown output
  -F, --frontmatter          Include frontmatter in markdown output (enabled by default for Hugo format)
  -M, --frontmatter-field=FIELD,...
                             Add custom frontmatter field (format: key:value), can be used multiple times
      --frontmatter-format="yaml"
                             Frontmatter format (yaml, toml, json)
  -t, --transport="command"  Transport type (command, sse, streamable)
      --endpoint=STRING      HTTP endpoint for SSE/Streamable transports
      --timeout=30s          HTTP timeout for SSE/Streamable transports
  -H, --headers=HEADERS,...  HTTP headers for SSE/Streamable transports (format: Key:Value)
      --context-file=CONTEXT-FILE,...
                             Path to context configuration files (YAML/JSON), can be used multiple times
      --server-command=STRING Server command for explicit command transport
      --no-tools             Skip scanning tools from the MCP server
      --no-resources         Skip scanning resources from the MCP server
      --no-prompts           Skip scanning prompts from the MCP server

Hugo-specific options (only used when format=hugo):
      --hugo-base-url=STRING           Base URL for Hugo site (e.g., https://example.com)
      --hugo-language-code=STRING      Language code for Hugo site (default: en-us)
      --hugo-enterprise-key=STRING     Enterprise key for Presidium configuration (optional, uses server name if not specified)
      --hugo-author-strict             Require author field in frontmatter (default: false)
      --custom-initialisms=STRING,...
                                       Additional technical initialisms to recognize for human-readable headings
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

### Rich Context Documentation Usage

```yaml
name: Generate Enhanced MCP Server Documentation with Context
on:
  push:
    branches: [ main ]

jobs:
  docs-with-context:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Generate Enhanced MCP Documentation
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'npx @modelcontextprotocol/server-filesystem ./data'
          format: 'html'
          output-file: 'docs/enhanced-server-docs.html'
          context-files: 'docs/server-context.yaml,docs/security-context.json'
          frontmatter: 'yaml'
          scan-tools: true      # Include tools (default)
          scan-resources: true  # Include resources (default)
          scan-prompts: false   # Skip prompts for faster generation

      - name: Upload enhanced documentation
        uses: actions/upload-artifact@v4
        with:
          name: enhanced-mcp-docs
          path: docs/enhanced-server-docs.html
```

**Context Files Configuration:**
- Context files are merged in the order specified (left to right)
- Later files override keys from earlier files
- Supports both YAML and JSON formats
- Files must be within the workspace directory for security
- Non-existent files are skipped with a warning
- See the [Rich Structured Context](#rich-structured-context) section for detailed configuration format

**Path Handling for Context Files:**
- **Relative paths** (recommended): `docs/context.yaml`, `config/server-context.json`
  - Resolved relative to the GitHub Action workspace (repository root)
  - More portable and secure for CI/CD environments
  - Example: `context-files: 'docs/context.yaml,config/overrides.json'`

- **Absolute paths**: Only if context files are outside the repository
  - Must be within the workspace directory for security
  - Not recommended for most use cases
  - Example: `context-files: '/workspace/external-config.yaml'`

- **Security restrictions**:
  - Path traversal attempts (`../`, `../../`) are blocked automatically
  - Files outside the workspace directory are inaccessible
  - Suspicious path patterns are logged and skipped

### Tool Calling Usage

```yaml
name: Generate MCP Documentation with Tool Calls
on:
  push:
    branches: [ main ]

jobs:
  docs-with-tool-calls:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Generate Documentation with Specific Tool Calls
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'node server.js'
          format: 'html'
          output-file: 'docs/server-with-tools.html'
          call-tool: 'get_weather,get_forecast'
          tool-args: '{"location":"London","units":"metric"}'

      - name: Generate Documentation with All Tools
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'node server.js'
          format: 'markdown'
          output-file: 'docs/server-all-tools.md'
          call-all-tools: 'true'

      - name: Upload documentation with tool results
        uses: actions/upload-artifact@v4
        with:
          name: mcp-docs-with-tools
          path: docs/server-*.html
```

**Tool Calling Configuration:**
- **call-tool**: Comma-separated list of tool names to call (e.g., `'tool1,tool2,tool3'`)
- **tool-args**: JSON string with arguments to pass to all tools (e.g., `'{"param":"value"}'`)
- **call-all-tools**: Set to `'true'` to call all available tools for comprehensive testing
- Tool call results are automatically included in the documentation output
- Failed tool calls are logged with error messages in the output
- Works with all output formats (markdown, html, json, pdf, hugo)

### Selective Scanning Usage

```yaml
name: Generate Focused MCP Documentation
on:
  push:
    branches: [ main ]

jobs:
  tools-only-docs:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Generate Tools-Only Documentation
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'node tool-server.js'
          format: 'markdown'
          output-file: 'docs/tools-only.md'
          scan-tools: true      # Include tools
          scan-resources: false # Skip resources for performance
          scan-prompts: false   # Skip prompts for performance

      - name: Generate Resources-Only Documentation
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'python resource-server.py'
          format: 'html'
          output-file: 'docs/resources-only.html'
          scan-tools: false     # Skip tools
          scan-resources: true  # Include resources
          scan-prompts: false   # Skip prompts

      - name: Upload focused documentation
        uses: actions/upload-artifact@v4
        with:
          name: focused-docs
          path: docs/
```

### Hugo Documentation Site Usage

```yaml
name: Generate Hugo Documentation Site
on:
  push:
    branches: [ main ]
  workflow_dispatch:

jobs:
  generate-hugo-site:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Generate Hugo Documentation Site
        uses: spandigital/mcp-server-dump@v1
        with:
          server-command: 'npx @modelcontextprotocol/server-filesystem ./data'
          format: 'hugo'
          output-file: 'docs-site'
          hugo-base-url: 'https://docs.mycompany.com'
          hugo-github: 'mycompany'
          hugo-twitter: 'mycompany'
          hugo-google-analytics: 'G-1234567890'
          hugo-site-logo: 'static/logo.svg'

      - name: Setup Hugo
        uses: peaceiris/actions-hugo@v3
        with:
          hugo-version: 'latest'
          extended: true

      - name: Build Hugo Site
        run: |
          cd docs-site
          hugo mod init docs.mycompany.com
          hugo mod get github.com/spandigital/presidium-styling-base
          hugo mod get github.com/spandigital/presidium-layouts-base
          hugo --minify

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        if: github.ref == 'refs/heads/main'
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs-site/public
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
| `format` | Output format (markdown, html, json, pdf, hugo) | No | `markdown` |
| `output-file` | Output file path (required for pdf format) or directory path (for hugo format) | No | - |
| `no-toc` | Disable table of contents in markdown output | No | `false` |
| `frontmatter` | Add frontmatter to output (yaml, toml, json) | No | - |
| `timeout` | Connection timeout in seconds | No | `30` |
| `verbose` | Enable verbose output | No | `false` |
| `context-files` | Context configuration files (YAML/JSON) for rich documentation (comma-separated). Files are merged in order, with later files overriding earlier ones. | No | - |
| `scan-tools` | Include tools in the documentation output | No | `true` |
| `scan-resources` | Include resources in the documentation output | No | `true` |
| `scan-prompts` | Include prompts in the documentation output | No | `true` |
| `hugo-base-url` | Base URL for Hugo site (e.g., https://example.com or https://docs.mysite.com) | No | - |
| `hugo-language-code` | Language code for Hugo site (e.g., en, en-US, zh-Hans, pt-BR) | No | `en-us` |
| `custom-initialisms` | Additional technical initialisms to recognize for human-readable headings (comma-separated, e.g., API,CDN,JWT,CORP,ACME) | No | - |

**Parameter Naming Convention**: The GitHub Action uses positive naming (`scan-tools: true/false`) while the CLI uses negative flags (`--no-tools`). This is intentional - GitHub Actions typically use positive boolean parameters for better UX, while CLI tools often use negative flags to disable default behavior.

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
