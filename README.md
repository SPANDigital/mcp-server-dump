# mcp-server-dump

A command-line tool to dump MCP (Model Context Protocol) server capabilities and documentation for reverse engineering purposes.

## Features

- **Multiple Transport Support**: Connect to MCP servers via various transports:
  - STDIO/Command transport (subprocess execution)
  - Server-Sent Events (SSE) over HTTP
  - Streamable HTTP transport
- Extract server information, capabilities, tools, resources, and prompts
- Output documentation in Markdown, JSON, or HTML format
- **Enhanced Markdown output with clickable Table of Contents**
- **External Go templates for customizable documentation**
- **Backward compatible** with existing command-line usage
- Built with the official [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- Clean CLI interface powered by [Kong](https://github.com/alecthomas/kong)

## SPANDigital and Private Repos

Make sure you have Github CLI

```bash
brew install gh
gh auth login
```
and have this is in your .zshrc or .bashrc

```bash
go env -w GOPRIVATE="github.com/spandigital/*"
echo "machine github.com login richardwooding password $(gh auth token)" > ~/.netrc
export PATH=$PATH:$(go env GOPATH)/bin
```

## Installation

### Using go install (Recommended)

```bash
go install github.com/SPANDigital/mcp-server-dump@latest
```

The binary will be installed to `$GOPATH/bin/mcp-server-dump` (or `$(go env GOPATH)/bin/mcp-server-dump`). Make sure your Go bin directory is in your PATH.

### From Source

```bash
git clone https://github.com/spandigtial/mcp-server-dump.git
cd mcp-server-dump
go build -o mcp-server-dump
```

### Requirements

- Go 1.25.0 or later
- Access to MCP servers (Node.js, Python, etc.)

## Usage

### Basic Usage

```bash
# Connect to a Node.js MCP server (default command transport)
./mcp-server-dump node server.js

# Connect to a Python MCP server with arguments
./mcp-server-dump python server.py --config config.json

# Connect to an NPX package
./mcp-server-dump npx @modelcontextprotocol/server-filesystem /path/to/directory
```

### Transport Options

```bash
# Command transport (default) - runs server as subprocess
./mcp-server-dump --transport=command node server.js
./mcp-server-dump --transport=command --server-command="python server.py --arg value"

# SSE transport - connects to HTTP Server-Sent Events endpoint
./mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse"

# Streamable transport - connects to HTTP streamable endpoint
./mcp-server-dump --transport=streamable --endpoint="http://localhost:3001/stream"

# Configure HTTP timeout for web transports
./mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse" --timeout=60s

# Add custom HTTP headers for authentication or other purposes
./mcp-server-dump --transport=sse --endpoint="http://localhost:3001/sse" \
  -H "Authorization:Bearer your-token-here" \
  -H "X-API-Key:your-api-key"

# Disable table of contents in markdown output
./mcp-server-dump --no-toc node server.js

# Generate HTML output from markdown
./mcp-server-dump -f html node server.js

# HTML output without table of contents
./mcp-server-dump -f html --no-toc node server.js
```

### Output Options

```bash
# Output to file (Markdown by default)
./mcp-server-dump -o server-docs.md node server.js

# JSON output
./mcp-server-dump -f json node server.js

# HTML output
./mcp-server-dump -f html node server.js

# Output to file (any format)
./mcp-server-dump -f json -o server-info.json python server.py
./mcp-server-dump -f html -o server-docs.html python server.py
```

### Command Line Options

```
Usage: mcp-server-dump [<args> ...] [flags]

Arguments:
  [<args> ...]               Command and arguments (legacy format for backward compatibility)

Flags:
  -h, --help                 Show context-sensitive help
  -o, --output=STRING        Output file for documentation (defaults to stdout)
  -f, --format="markdown"    Output format (markdown, json, html)
      --no-toc               Disable table of contents in markdown output
  -t, --transport="command"  Transport type (command, sse, streamable)
      --endpoint=STRING      HTTP endpoint for SSE/Streamable transports
      --timeout=30s          HTTP timeout for SSE/Streamable transports
  -H, --headers=HEADERS,...  HTTP headers for SSE/Streamable transports (format: Key:Value)
      --server-command=STRING Server command for explicit command transport
```

## Example Output

### Markdown Format (Enhanced with Table of Contents)
```markdown
# example-server

**Version:** 1.0.0

## Table of Contents

- [Capabilities](#capabilities)
- [Tools](#tools)
  - [read_file](#tool-read-file)
  - [write_file](#tool-write-file)
- [Resources](#resources)
  - [example.txt](#resource-example-txt)

## Capabilities {#capabilities}

- **Tools:** ✅ Supported
- **Resources:** ✅ Supported
- **Prompts:** ❌ Not supported

## Tools {#tools}

### read_file {#tool-read-file}

Read contents of a file from the filesystem

**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to the file to read"
    }
  }
}
```

## Resources {#resources}

### example.txt {#resource-example-txt}

**URI:** `file://example.txt`

Example text file resource

**MIME Type:** text/plain
```

### HTML Format (GitHub Flavored Markdown Compatible)

HTML output is generated by converting the Markdown output using [Goldmark](https://github.com/yuin/goldmark) with GitHub Flavored Markdown extensions:

```html
<h1 id="example-server">example-server</h1>
<p><strong>Version:</strong> 1.0.0</p>

<h2 id="table-of-contents">Table of Contents</h2>
<ul>
<li><a href="#capabilities">Capabilities</a></li>
<li><a href="#tools">Tools</a>
<ul>
<li><a href="#read-file">read_file</a></li>
<li><a href="#write-file">write_file</a></li>
</ul>
</li>
</ul>

<h2 id="capabilities">Capabilities</h2>
<ul>
<li><strong>Tools:</strong> ✅ Supported</li>
<li><strong>Resources:</strong> ✅ Supported</li>
<li><strong>Prompts:</strong> ❌ Not supported</li>
</ul>

<h2 id="tools">Tools</h2>
<h3 id="read-file">read_file</h3>
<p>Read contents of a file from the filesystem</p>
<p><strong>Input Schema:</strong></p>
<pre><code class="language-json">{
  &quot;type&quot;: &quot;object&quot;,
  &quot;properties&quot;: {
    &quot;path&quot;: {
      &quot;type&quot;: &quot;string&quot;,
      &quot;description&quot;: &quot;Path to the file to read&quot;
    }
  }
}
</code></pre>
```

### JSON Format
```json
{
  "name": "example-server",
  "version": "1.0.0",
  "capabilities": {
    "tools": true,
    "resources": true,
    "prompts": false
  },
  "tools": [
    {
      "name": "read_file",
      "description": "Read contents of a file from the filesystem",
      "inputSchema": {
        "type": "object",
        "properties": {
          "path": {
            "type": "string",
            "description": "Path to the file to read"
          }
        }
      }
    }
  ],
  "resources": [
    {
      "uri": "file://example.txt",
      "name": "example.txt",
      "description": "Example text file resource",
      "mimeType": "text/plain"
    }
  ],
  "prompts": []
}
```

## Use Cases

- **Server Documentation**: Generate comprehensive documentation for MCP servers
- **API Discovery**: Understand available tools and resources in MCP servers
- **Integration Planning**: Analyze server capabilities before integration
- **Reverse Engineering**: Explore and document third-party MCP servers
- **Testing & Validation**: Verify server implementations match specifications

## Dependencies

- [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - Official MCP Go SDK
- [github.com/alecthomas/kong](https://github.com/alecthomas/kong) - Command line parser

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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details.
