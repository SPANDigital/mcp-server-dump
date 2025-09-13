# Contributing to mcp-server-dump

Thank you for your interest in contributing to mcp-server-dump! We welcome contributions from the community and are grateful for your support.

## Code of Conduct

Please note that this project is released with a [Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## How to Contribute

### Reporting Issues

Before creating an issue, please check if it has already been reported:
- Search through [existing issues](https://github.com/SPANDigital/mcp-server-dump/issues)
- Check both open and closed issues

When creating a new issue, please include:
- A clear, descriptive title
- Steps to reproduce the problem
- Expected behavior vs actual behavior
- Your environment (OS, Go version, etc.)
- Any relevant error messages or logs

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:
- A clear and descriptive title
- A detailed description of the proposed enhancement
- Examples of how the feature would be used
- Any potential drawbacks or considerations

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Follow the existing code style** - run `go fmt` and `go vet`
3. **Write tests** for new functionality when applicable
4. **Update documentation** - update README.md and CLAUDE.md if needed
5. **Test your changes** thoroughly with different MCP servers
6. **Create a Pull Request** with a clear description

#### Development Setup

```bash
# Clone your fork
git clone https://github.com/your-username/mcp-server-dump.git
cd mcp-server-dump

# Install dependencies
go mod download

# Build the project
go build -o mcp-server-dump ./cmd/mcp-server-dump

# Run tests
go test ./...

# Run linter (install golangci-lint first)
golangci-lint run
```

#### Commit Messages

Please follow these guidelines for commit messages:
- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests when relevant

Common prefixes:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Test additions or changes
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

#### Testing

- Write unit tests for new functionality
- Test with multiple MCP server implementations
- Verify all output formats work correctly (Markdown, JSON, HTML, PDF)
- Test different transport types (stdio, SSE, streamable)

### Code Style

- Follow standard Go conventions
- Use `go fmt` to format your code
- Run `go vet` to catch common issues
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and concise

### Documentation

- Update README.md for user-facing changes
- Update CLAUDE.md for architectural changes
- Include examples for new features
- Keep documentation clear and concise

## Project Structure

```
cmd/mcp-server-dump/    - Entry point
internal/
├── app/                - Application logic
├── transport/          - MCP transport implementations
├── formatter/          - Output formatters
└── model/             - Data structures
```

## Release Process

Releases are automated via GitHub Actions when a new tag is pushed. Maintainers will handle the release process.

## Getting Help

If you need help, you can:
- Open a [GitHub issue](https://github.com/SPANDigital/mcp-server-dump/issues)
- Check the [README](README.md) for usage information
- Review existing issues and pull requests

## Recognition

Contributors will be recognized in the project's release notes. Thank you for helping make mcp-server-dump better!

## License

By contributing to mcp-server-dump, you agree that your contributions will be licensed under the MIT License.