# Linux Package Repository Setup

This document describes how to use the self-hosted APT and YUM repositories for installing mcp-server-dump on Linux systems.

## Repository URLs

The Linux packages are automatically published to GitHub Pages after each release:

- **Base URL**: https://spandigital.github.io/mcp-server-dump/
- **APT Repository**: https://spandigital.github.io/mcp-server-dump/apt
- **YUM Repository**: https://spandigital.github.io/mcp-server-dump/yum

## Installation Instructions

⚠️ **Security Notice**: The repository currently uses simplified configuration for ease of setup. For production environments, consider enabling GPG verification when available.

### Debian/Ubuntu (APT)

#### Quick Install (Unsigned)
```bash
# Add the repository (without GPG verification - not recommended for production)
echo "deb [trusted=yes] https://spandigital.github.io/mcp-server-dump/apt stable main" | sudo tee /etc/apt/sources.list.d/mcp-server-dump.list

# Update package list and install
sudo apt update
sudo apt install mcp-server-dump
```

#### Secure Install (GPG Verified) - Recommended
```bash
# Import GPG key (when available)
curl -fsSL https://spandigital.github.io/mcp-server-dump/public.key | sudo gpg --dearmor -o /usr/share/keyrings/mcp-server-dump.gpg

# Add the repository with GPG verification
echo "deb [signed-by=/usr/share/keyrings/mcp-server-dump.gpg] https://spandigital.github.io/mcp-server-dump/apt stable main" | sudo tee /etc/apt/sources.list.d/mcp-server-dump.list

# Update package list and install
sudo apt update
sudo apt install mcp-server-dump
```

### RHEL/Fedora/CentOS (YUM/DNF)

#### Quick Install (Unsigned)
```bash
# Create repository configuration (without GPG verification - not recommended for production)
sudo tee /etc/yum.repos.d/mcp-server-dump.repo << 'EOF'
[mcp-server-dump]
name=MCP Server Dump
baseurl=https://spandigital.github.io/mcp-server-dump/yum/$basearch
enabled=1
gpgcheck=0
EOF

# Install the package
sudo dnf install mcp-server-dump  # For Fedora/RHEL 8+
# OR
sudo yum install mcp-server-dump  # For older RHEL/CentOS
```

#### Secure Install (GPG Verified) - Recommended
```bash
# Import GPG key (when available)
sudo rpm --import https://spandigital.github.io/mcp-server-dump/public.key

# Create repository configuration with GPG verification
sudo tee /etc/yum.repos.d/mcp-server-dump.repo << 'EOF'
[mcp-server-dump]
name=MCP Server Dump
baseurl=https://spandigital.github.io/mcp-server-dump/yum/$basearch
enabled=1
gpgcheck=1
gpgkey=https://spandigital.github.io/mcp-server-dump/public.key
EOF

# Install the package
sudo dnf install mcp-server-dump  # For Fedora/RHEL 8+
# OR
sudo yum install mcp-server-dump  # For older RHEL/CentOS
```

## Direct Package Download

If you prefer to download packages directly, they are available as release assets:

- Visit [Releases](https://github.com/spandigital/mcp-server-dump/releases/latest)
- Download the appropriate package for your distribution:
  - `.deb` for Debian/Ubuntu
  - `.rpm` for RHEL/Fedora/CentOS
  - `.apk` for Alpine Linux

### Manual Installation

```bash
# Debian/Ubuntu
sudo dpkg -i mcp-server-dump_*.deb
# OR
sudo apt install ./mcp-server-dump_*.deb

# RHEL/Fedora/CentOS
sudo rpm -i mcp-server-dump-*.rpm
# OR
sudo dnf install ./mcp-server-dump-*.rpm

# Alpine Linux
sudo apk add --allow-untrusted mcp-server-dump_*.apk
```

## Updating

Once the repository is configured, updates will be available through your package manager:

```bash
# Debian/Ubuntu
sudo apt update && sudo apt upgrade mcp-server-dump

# RHEL/Fedora/CentOS
sudo dnf upgrade mcp-server-dump
```

## Uninstalling

```bash
# Debian/Ubuntu
sudo apt remove mcp-server-dump

# RHEL/Fedora/CentOS
sudo dnf remove mcp-server-dump

# Alpine Linux
sudo apk del mcp-server-dump
```

## Repository Structure

The repository is automatically maintained by GitHub Actions and includes:

- `/apt/` - APT repository for Debian-based distributions
  - `/pool/` - Package pool containing .deb files
  - `/dists/stable/` - Distribution metadata and indices
- `/yum/` - YUM repository for RPM-based distributions
  - `/x86_64/` - x86_64 architecture packages
  - `/aarch64/` - ARM64 architecture packages

## Security Considerations

### Package Signing Status
The repository infrastructure supports GPG signing, but packages may be unsigned in initial releases for simplicity. Check the repository status page for current signing status.

### Recommended Security Practices
1. **Use GPG verification when available**: Prefer the secure installation methods shown above
2. **Verify package integrity**: Compare SHA256 checksums when downloading packages directly
3. **Repository security**: The repository is hosted on GitHub Pages with TLS encryption
4. **Source verification**: All packages are built from the official [GitHub repository](https://github.com/spandigital/mcp-server-dump)

### Package Verification
When GPG signing is enabled, you can verify packages manually:
```bash
# For DEB packages
gpg --verify package.deb.asc package.deb

# For RPM packages
rpm --checksig package.rpm
```

### Trust Considerations
- Packages are built automatically from tagged releases using GitHub Actions
- The build process is auditable via the public [workflow logs](https://github.com/spandigital/mcp-server-dump/actions)
- Repository hosting uses GitHub's infrastructure security

## Troubleshooting

### APT Issues

If you encounter issues with the APT repository:

```bash
# Clear APT cache
sudo apt clean
sudo apt update

# Remove and re-add the repository
sudo rm /etc/apt/sources.list.d/mcp-server-dump.list
# Then follow the installation instructions again
```

### YUM/DNF Issues

If you encounter issues with the YUM repository:

```bash
# Clear YUM cache
sudo dnf clean all  # or yum clean all

# Remove and re-add the repository
sudo rm /etc/yum.repos.d/mcp-server-dump.repo
# Then follow the installation instructions again
```

## Contributing

The repository publishing is automated through GitHub Actions. See `.github/workflows/publish-repos.yml` for the implementation details.