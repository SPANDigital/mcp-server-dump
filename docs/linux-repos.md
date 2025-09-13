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

#### Recommended: Secure Install (GPG Verified)
```bash
# Import GPG key
curl -fsSL https://spandigital.github.io/mcp-server-dump/public.key | sudo gpg --dearmor -o /usr/share/keyrings/mcp-server-dump.gpg

# Add the repository with GPG verification
echo "deb [signed-by=/usr/share/keyrings/mcp-server-dump.gpg] https://spandigital.github.io/mcp-server-dump/apt stable main" | sudo tee /etc/apt/sources.list.d/mcp-server-dump.list

# Update package list and install
sudo apt update
sudo apt install mcp-server-dump
```

#### Alternative: Quick Install (Unsigned - Development/Testing Only)
⚠️ **WARNING**: This method bypasses package signature verification and should only be used for development or testing environments.

```bash
# Add the repository (without GPG verification - NOT RECOMMENDED for production)
echo "deb [trusted=yes] https://spandigital.github.io/mcp-server-dump/apt stable main" | sudo tee /etc/apt/sources.list.d/mcp-server-dump.list

# Update package list and install
sudo apt update
sudo apt install mcp-server-dump
```

### RHEL/Fedora/CentOS (YUM/DNF)

#### Recommended: Secure Install (GPG Verified)
```bash
# Import GPG key
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

#### Alternative: Quick Install (Unsigned - Development/Testing Only)
⚠️ **WARNING**: This method bypasses package signature verification and should only be used for development or testing environments.

```bash
# Create repository configuration (without GPG verification - NOT RECOMMENDED for production)
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

### Alpine Linux (APK)

Alpine Linux packages are available as direct downloads from GitHub releases. APK packages are not hosted in a repository due to Alpine's package signing requirements.

#### Manual Installation
```bash
# Download the APK package from the latest release
wget https://github.com/spandigital/mcp-server-dump/releases/latest/download/mcp-server-dump_linux_amd64.apk

# Install the package (requires --allow-untrusted for non-official packages)
sudo apk add --allow-untrusted mcp-server-dump_linux_amd64.apk
```

#### Using Alpine Package Manager with Direct URL
```bash
# Install directly from URL
sudo apk add --allow-untrusted https://github.com/spandigital/mcp-server-dump/releases/latest/download/mcp-server-dump_linux_amd64.apk
```

**Note**: Replace `amd64` with `arm64` for ARM-based systems.

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

### GPG Key Management Strategy
The project uses a dedicated GPG key for package signing:

- **Key ID**: `9079FAE841B09114`
- **Key Type**: 4096-bit RSA
- **Expires**: September 13, 2027
- **Owner**: SPAN Digital <richard.wooding@spandigital.com>
- **Purpose**: Linux package and repository signing only

**Key Storage**:
- Private key stored as GitHub repository secret (`GPG_PRIVATE_KEY`)
- Public key available in repository (`docs/public.key`) and published to GitHub Pages
- Key rotation planned before expiration with advance notice

**Private Key Backup & Recovery**:
- Private key is securely backed up offline by project maintainers
- Recovery process documented in internal security procedures
- Emergency contact: richard.wooding@spandigital.com for key-related issues
- Backup verification performed annually

**Key Expiration Monitoring**:
- GPG key expires September 13, 2027 (3 years from creation)
- Automated monitoring planned for 6 months before expiration
- Key rotation will be performed with 3 months advance notice
- New key will be cross-signed by the expiring key for continuity

**Key Verification**:
```bash
# Download and verify the public key
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/public.key | gpg --import
gpg --fingerprint 9079FAE841B09114
# Expected fingerprint: 3340 5835 CE12 B5D3 440F 1611 9079 FAE8 41B0 9114
```

### Trust Considerations
- Packages are built automatically from tagged releases using GitHub Actions
- The build process is auditable via the public [workflow logs](https://github.com/spandigital/mcp-server-dump/actions)
- Repository hosting uses GitHub's infrastructure security
- GPG signing provides cryptographic verification of package authenticity
- Key management follows security best practices with time-limited keys

### Supply Chain Security
- All packages include SHA256 checksums for integrity verification
- Release assets are cryptographically signed using GPG
- Build process uses reproducible builds where possible
- Dependency updates are tracked and auditable via GitHub
- Package authenticity can be verified through multiple methods:
  ```bash
  # Verify GPG signature
  gpg --verify package.sig package.deb

  # Verify SHA256 checksum
  sha256sum -c package.sha256
  ```

### Repository Performance & Rate Limiting

**GitHub Pages Hosting**:
- Repositories are hosted on GitHub Pages with global CDN distribution
- No explicit rate limits for package downloads from GitHub Pages
- Best practices for automated systems:
  - Implement reasonable delays between package manager operations
  - Use caching where possible to reduce repository metadata requests
  - Consider mirroring for high-volume enterprise deployments

**Repository Mirrors**:
For high-availability deployments, consider setting up repository mirrors:
```bash
# Example: Sync repository to internal mirror
rsync -av --delete https://spandigital.github.io/mcp-server-dump/ /local/mirror/
```

**Enterprise Considerations**:
- Large organizations may want to cache packages internally
- Consider using tools like Nexus Repository or JFrog Artifactory for mirroring
- Monitor repository access patterns and implement caching strategies

## Troubleshooting

### GPG Issues

**GPG Key Import Failures:**
```bash
# Error: "gpg: no valid OpenPGP data found"
# Solution: Re-download the key with proper error checking
curl -fsSL https://spandigital.github.io/mcp-server-dump/public.key -o /tmp/mcp-key.asc
file /tmp/mcp-key.asc  # Should show "PGP public key block"
sudo gpg --dearmor /tmp/mcp-key.asc -o /usr/share/keyrings/mcp-server-dump.gpg
```

**Signature Verification Failures:**
```bash
# Error: "signatures were invalid" or "NO_PUBKEY"
# Solution: Re-import the GPG key and update package cache
sudo rm -f /usr/share/keyrings/mcp-server-dump.gpg
curl -fsSL https://spandigital.github.io/mcp-server-dump/public.key | sudo gpg --dearmor -o /usr/share/keyrings/mcp-server-dump.gpg
sudo apt update
```

**RPM GPG Key Trust Issues:**
```bash
# Error: "NOKEY" or "Public key is not installed"
# Solution: Re-import the RPM GPG key
sudo rpm --import https://spandigital.github.io/mcp-server-dump/public.key
# Verify key is imported
rpm -qa gpg-pubkey --qf '%{name}-%{version}-%{release} --> %{summary}\n'
```

**Key Fingerprint Verification:**
```bash
# Always verify the key fingerprint matches the expected value
gpg --show-keys /usr/share/keyrings/mcp-server-dump.gpg
# Expected fingerprint: 3340 5835 CE12 B5D3 440F 1611 9079 FAE8 41B0 9114
```

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

## Repository Maintenance

### Package Retention Policy

The Linux package repositories grow with each release as new packages are added. To manage repository size:

**Current Strategy:**
- All package versions are retained indefinitely
- Repository metadata is regenerated for each release
- Old packages remain accessible for users who need specific versions

**Future Considerations:**
As the repository grows, maintainers may implement:
- Retention of last N major versions (e.g., keep last 10 releases)
- Automated cleanup of packages older than X months
- Separate archive repository for historical packages

**Manual Cleanup (Maintainers Only):**
```bash
# To manually clean up old packages from linux-repos branch:
git checkout linux-repos
find apt/pool -name "*.deb" -mtime +365 -delete  # Remove packages older than 1 year
find yum -name "*.rpm" -mtime +365 -delete       # Remove RPM packages older than 1 year
# Regenerate repository metadata after cleanup
```

### Repository Health Monitoring

**Key Metrics to Monitor:**
- Repository size growth rate
- Package download statistics
- GPG key expiration (current key expires September 13, 2027)
- Failed repository updates or publishing errors

**Health Checks:**
- Verify repository metadata integrity monthly
- Test package installation from repositories quarterly
- Monitor GitHub Pages bandwidth usage

## Contributing

The repository publishing is automated through GitHub Actions. See `.github/workflows/publish-repos.yml` for the implementation details.