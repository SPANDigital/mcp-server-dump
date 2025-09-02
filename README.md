# mcp-server-dump

A command-line tool to dump MCP (Model Context Protocol) server capabilities and documentation for reverse engineering purposes.

## Features

- Connect to MCP servers via STDIO (command execution)
- Extract server information, capabilities, tools, resources, and prompts
- Output documentation in Markdown or JSON format
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
export GITHUB_TOKEN=$(gh auth token)
export GITHUB_EMAIL=richard.wooding@spandigital.com
export GITHUB_USER="Richard Wooding"
go env -w GOPRIVATE="github.com/spandigital/*"
echo "machine github.com login richardwooding password ${GITHUB_TOKEN}" > ~/.netrc
export PATH=$PATH:$(go env GOPATH)/bin
```

## Installation

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
# Connect to a Node.js MCP server
./mcp-server-dump node server.js

# Connect to a Python MCP server with arguments
./mcp-server-dump python server.py --config config.json

# Connect to an NPX package
./mcp-server-dump npx @modelcontextprotocol/server-filesystem /path/to/directory
```

### Output Options

```bash
# Output to file (Markdown by default)
./mcp-server-dump -o server-docs.md node server.js

# JSON output
./mcp-server-dump -f json node server.js

# JSON output to file
./mcp-server-dump -f json -o server-info.json python server.py
```

### Command Line Options

```
Usage: mcp-server-dump <command> [<args> ...] [flags]

Arguments:
  <command>       Command to execute the MCP server
  [<args> ...]    Arguments to pass to the MCP server command

Flags:
  -h, --help                 Show context-sensitive help
  -o, --output=STRING        Output file for documentation (defaults to stdout)
  -f, --format="markdown"    Output format (markdown, json)
```

## Example Output

### Markdown Format
```markdown
# example-server

**Version:** 1.0.0

## Capabilities

- Tools: true
- Resources: true
- Prompts: false

## Tools

### read_file

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

## Resources

### file://example.txt

**URI:** `file://example.txt`

Example text file resource

**MIME Type:** text/plain
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

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details.
