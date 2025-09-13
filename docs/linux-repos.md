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

- **Key ID**: `9079FAE841B09114` (configurable via repository variables)
- **Key Type**: 4096-bit RSA
- **Expires**: September 13, 2027
- **Owner**: SPAN Digital <richard.wooding@spandigital.com>
- **Purpose**: Linux package and repository signing only

**Key ID Configuration**:
The GPG key ID can be configured via GitHub repository variables (`GPG_KEY_ID`) to support fork-friendly development and key rotation.

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
- **Automated monitoring implemented**: Monthly checks via GitHub Actions workflow
- Key rotation alerts:
  - **180 days**: Medium priority notice for planning
  - **90 days**: High priority alert to begin rotation
  - **30 days**: Critical alert requiring immediate action
- Key rotation will be performed with 3 months advance notice
- New key will be cross-signed by the expiring key for continuity

**Key Verification**:
```bash
# Download and verify the public key
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/public.key | gpg --import
gpg --fingerprint 9079FAE841B09114
# Expected fingerprint: 3340 5835 CE12 B5D3 440F 1611 9079 FAE8 41B0 9114
```

### GPG Key Rotation Process

**When Key Rotation is Required**:
- GPG key expiration (current key expires September 13, 2027)
- Key compromise or suspected security breach
- Cryptographic algorithm deprecation
- Planned security improvements

**Key Rotation Timeline**:
1. **6 months before expiration**: Planning and preparation phase
2. **3 months before expiration**: New key generation and cross-signing
3. **2 months before expiration**: Testing and documentation updates
4. **1 month before expiration**: User notification and transition announcement
5. **Key expiration**: Old key becomes invalid

**Detailed Rotation Steps**:

**Phase 1: New Key Generation (3 months before expiry)**
```bash
# 1. Generate new GPG key with same parameters
gpg --batch --full-generate-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: SPAN Digital
Name-Email: richard.wooding@spandigital.com
Expire-Date: 3y
%commit
EOF

# 2. Export new keys
NEW_KEY_ID=$(gpg --list-keys --keyid-format LONG | grep 'pub.*rsa4096' | tail -n1 | cut -d'/' -f2 | cut -d' ' -f1)
gpg --armor --export-secret-keys $NEW_KEY_ID > new-private-key.asc
gpg --armor --export $NEW_KEY_ID > new-public-key.asc

# 3. Cross-sign with existing key for trust continuity
gpg --sign-key $NEW_KEY_ID
```

**Phase 2: Cross-Signing and Validation (2-3 months before expiry)**
```bash
# 1. Sign new key with old key
gpg --local-user 9079FAE841B09114 --sign-key $NEW_KEY_ID

# 2. Create transition announcement signature
echo "GPG Key Transition Notice for SPAN Digital mcp-server-dump" | \
  gpg --clearsign --local-user 9079FAE841B09114 > transition-notice.asc

# 3. Test new key with package signing
gpg --detach-sign --armor --local-user $NEW_KEY_ID test-package.deb
gpg --verify test-package.deb.asc test-package.deb
```

**Phase 3: Repository Updates (1-2 months before expiry)**
```bash
# 1. Update GitHub repository secrets
# (Done via GitHub web interface or API)
# - Update GPG_PRIVATE_KEY secret with new private key
# - Update GPG_PUBLIC_KEY secret with new public key

# 2. Update public key file in repository
cp new-public-key.asc docs/public.key
git add docs/public.key
git commit -m "docs: Update GPG public key for package signing"

# 3. Update documentation with new key ID and fingerprint
# Edit docs/linux-repos.md with new key details
```

**Phase 4: User Communication (1 month before expiry)**
```bash
# 1. Create GitHub release with transition notice
gh release create v1.x.x-key-transition --title "GPG Key Transition Notice" \
  --notes "This release includes a GPG key transition. Please update your package verification..."

# 2. Update installation documentation
# Ensure both old and new keys are documented during transition period

# 3. Notify users via multiple channels
# - GitHub release announcement
# - Repository README update
# - Issue tracker notification
```

**Phase 5: Key Transition Completion (after successful transition)**
```bash
# 1. Revoke old key (only after confirming new key works)
gpg --gen-revoke 9079FAE841B09114 > old-key-revocation.asc
gpg --import old-key-revocation.asc

# 2. Publish revocation
gpg --send-keys 9079FAE841B09114

# 3. Update all documentation to reference only new key
# 4. Clean up transition materials after 6 months
```

**User Instructions During Transition**:
```bash
# Users can import both keys during transition period
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/public.key | gpg --import
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/old-public.key | gpg --import

# Verify packages can be verified with either key
gpg --verify package.sig package.deb
```

**Emergency Key Rotation**:
In case of key compromise, the rotation process is accelerated:
1. **Immediate**: Generate new key and update repository secrets
2. **Within 24 hours**: Release new packages with new key
3. **Within 48 hours**: Revoke compromised key and notify users
4. **Within 1 week**: Complete documentation updates and user communication

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

**Security Warning - Expired Key:**
```bash
# Error: "gpg: key 9079FAE841B09114: signature expired"
# This indicates the GPG key has expired (current key expires September 13, 2027)
# Solution: Check for key rotation announcements and import new key

# 1. Check GitHub releases for key transition notices
# 2. Import new public key if available
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/public.key | gpg --import

# 3. If no new key is available, report security issue immediately
# Do NOT bypass signature verification with [trusted=yes] in production
```

**Key Transition Handling:**
```bash
# During key rotation periods, both old and new keys may be valid
# Import all available keys for maximum compatibility
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/public.key | gpg --import
curl -fsSL https://spandigital.github.io/mcp-server-dump/docs/old-public.key | gpg --import 2>/dev/null || true

# Verify which key was used to sign packages
gpg --verify package.sig package.deb
```

**Advanced Security Verification:**
```bash
# For high-security environments, verify package authenticity through multiple channels
# 1. GPG signature verification (primary)
gpg --verify package.sig package.deb

# 2. SHA256 checksum verification (secondary)
sha256sum -c package.sha256

# 3. Build reproducibility verification (advanced)
# Download source code and verify build matches released package
# git clone https://github.com/spandigital/mcp-server-dump.git
# cd mcp-server-dump && git checkout <release-tag>
# goreleaser build --single-target --snapshot
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
- **Automated GPG key expiration monitoring** (implemented via GitHub Actions)

## Security Best Practices

### For Package Maintainers

**GPG Key Security:**
- Store private keys securely with backup procedures
- Use hardware security modules (HSM) for high-value keys in production
- Implement key rotation procedures with advance notice
- Monitor key expiration dates with automated alerts
- Cross-sign new keys with existing keys during rotation

**Repository Security:**
- Use GPG signing for all packages and repository metadata
- Implement multi-person authorization for key operations
- Audit package build processes regularly
- Monitor for unauthorized repository modifications
- Maintain reproducible builds where possible

**Access Control:**
- Restrict repository write access to necessary personnel only
- Use branch protection rules for repository branches
- Implement review processes for package updates
- Monitor GitHub Actions workflow logs
- Use environment-specific secrets for different deployment stages

### For Package Users

**Installation Security:**
- **Always verify GPG signatures** - Never use `[trusted=yes]` or `gpgcheck=0` in production
- Verify key fingerprints match expected values
- Import keys from official sources only
- Monitor for key transition announcements
- Report suspicious packages or signature failures immediately

**Ongoing Security:**
- Keep systems updated with latest packages
- Monitor security advisories for package vulnerabilities
- Use package pinning for critical systems
- Implement package integrity monitoring
- Regular security audits of installed packages

**Enterprise Considerations:**
- Mirror repositories internally for air-gapped environments
- Implement additional signature verification layers
- Use vulnerability scanning tools on packages
- Maintain incident response procedures for compromised packages
- Document package approval processes

### Security Contact

**Reporting Security Issues:**
- **Email**: richard.wooding@spandigital.com
- **Subject**: [SECURITY] mcp-server-dump package security issue
- **Response Time**: Within 24 hours for critical issues

**Security Incident Response:**
1. **Report received**: Acknowledgment within 24 hours
2. **Initial assessment**: Risk evaluation within 48 hours
3. **Fix development**: Coordinated response with timeline
4. **Public disclosure**: After fix is available and tested
5. **Post-incident review**: Process improvement documentation

## Contributing

The repository publishing is automated through GitHub Actions. See `.github/workflows/publish-repos.yml` for the implementation details.

### Repository Configuration

For forks or key rotation, configure these repository variables and secrets:

**Required Secrets**:
- `GPG_PRIVATE_KEY`: GPG private key for signing packages (armored format)
- `GPG_PUBLIC_KEY`: GPG public key for verification (armored format)

**Optional Variables**:
- `GPG_KEY_ID`: GPG key ID for signing (defaults to `9079FAE841B09114`)

The GPG key monitoring workflow automatically adapts to the configured key ID.