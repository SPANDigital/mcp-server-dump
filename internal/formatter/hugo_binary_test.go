package formatter

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const windowsOS = "windows"

// HugoBinaryTestHelper manages downloading and testing with official Hugo binary
type HugoBinaryTestHelper struct {
	Version     string
	BinaryPath  string
	TempDir     string
	DownloadURL string
	t           *testing.T
}

// NewHugoBinaryTestHelper creates a new helper for testing with Hugo binary
func NewHugoBinaryTestHelper(t *testing.T) *HugoBinaryTestHelper {
	t.Helper()

	// Use Hugo v0.150.1 (latest as of January 2025 - consider updating periodically)
	version := "0.150.1"

	// Determine architecture and OS for download URL
	goarch := runtime.GOARCH
	goos := runtime.GOOS

	// Map Go OS and arch to Hugo release names (using actual release naming)
	var hugoBinaryPlatform string
	switch goos {
	case "darwin":
		hugoBinaryPlatform = "darwin-universal" // Hugo uses universal binaries for macOS
	case windowsOS:
		if goarch == "amd64" {
			hugoBinaryPlatform = "windows-amd64"
		} else if goarch == "386" {
			hugoBinaryPlatform = "windows-386"
		} else {
			hugoBinaryPlatform = fmt.Sprintf("windows-%s", goarch)
		}
	case "linux":
		if goarch == "amd64" {
			hugoBinaryPlatform = "linux-amd64"
		} else if goarch == "arm64" {
			hugoBinaryPlatform = "linux-arm64"
		} else if goarch == "386" {
			hugoBinaryPlatform = "linux-386"
		} else {
			hugoBinaryPlatform = fmt.Sprintf("linux-%s", goarch)
		}
	default:
		hugoBinaryPlatform = fmt.Sprintf("%s-%s", goos, goarch)
	}

	// Construct download URL for Hugo extended (required for Sass/SCSS)
	var filename string
	if goos == windowsOS {
		filename = fmt.Sprintf("hugo_extended_%s_%s.zip", version, hugoBinaryPlatform)
	} else {
		filename = fmt.Sprintf("hugo_extended_%s_%s.tar.gz", version, hugoBinaryPlatform)
	}
	downloadURL := fmt.Sprintf("https://github.com/gohugoio/hugo/releases/download/v%s/%s", version, filename)

	// Create temporary directory for Hugo binary
	tempDir, err := os.MkdirTemp("", "hugo_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory for Hugo binary: %v", err)
	}

	binaryName := "hugo"
	if goos == windowsOS {
		binaryName = "hugo.exe"
	}
	binaryPath := filepath.Join(tempDir, binaryName)

	helper := &HugoBinaryTestHelper{
		Version:     version,
		BinaryPath:  binaryPath,
		TempDir:     tempDir,
		DownloadURL: downloadURL,
		t:           t,
	}

	// Use t.Cleanup for better resource management
	t.Cleanup(func() {
		helper.Cleanup()
	})

	return helper
}

// DownloadAndExtract downloads the Hugo binary and extracts it
func (h *HugoBinaryTestHelper) DownloadAndExtract() error {
	h.t.Helper()
	h.t.Logf("Downloading Hugo %s from %s", h.Version, h.DownloadURL)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Minute, // Large timeout for downloading
	}

	// Download the archive
	resp, err := client.Get(h.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download Hugo binary: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.t.Logf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Hugo binary: HTTP %d", resp.StatusCode)
	}

	// Extract based on file type
	if strings.HasSuffix(h.DownloadURL, ".tar.gz") {
		return h.extractTarGz(resp.Body)
	}

	return fmt.Errorf("unsupported archive format for URL: %s", h.DownloadURL)
}

// extractTarGz extracts a tar.gz archive containing Hugo binary
func (h *HugoBinaryTestHelper) extractTarGz(r io.Reader) error {
	// Decompress gzip
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if err := gzr.Close(); err != nil {
			h.t.Logf("Failed to close gzip reader: %v", err)
		}
	}()

	// Read tar archive
	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Look for hugo binary (might be in subdirectory)
		if strings.HasSuffix(header.Name, "hugo") || strings.HasSuffix(header.Name, "hugo.exe") {
			// Create the binary file
			outFile, err := os.OpenFile(h.BinaryPath, os.O_CREATE|os.O_WRONLY, 0o755)
			if err != nil {
				return fmt.Errorf("failed to create Hugo binary file: %w", err)
			}

			// Copy binary content
			_, copyErr := io.Copy(outFile, tr)
			if closeErr := outFile.Close(); closeErr != nil {
				return fmt.Errorf("failed to close Hugo binary file: %w", closeErr)
			}
			if copyErr != nil {
				return fmt.Errorf("failed to extract Hugo binary: %w", copyErr)
			}

			h.t.Logf("Hugo binary extracted to: %s", h.BinaryPath)
			return nil
		}
	}

	return fmt.Errorf("hugo binary not found in archive")
}

// RunHugo executes Hugo with given arguments and returns output
func (h *HugoBinaryTestHelper) RunHugo(args ...string) (string, error) {
	h.t.Helper()

	// Check if binary exists
	if _, err := os.Stat(h.BinaryPath); os.IsNotExist(err) {
		return "", fmt.Errorf("Hugo binary not found at %s (call DownloadAndExtract first)", h.BinaryPath)
	}

	h.t.Logf("Running: %s %v", h.BinaryPath, args)

	// Execute Hugo command using exec.Command
	cmd := exec.Command(h.BinaryPath, args...)

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("Hugo command failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// GetVersion returns Hugo version by executing hugo version
func (h *HugoBinaryTestHelper) GetVersion() (string, error) {
	h.t.Helper()
	return h.RunHugo("version")
}

// InitModule initializes a Hugo module in the given directory
func (h *HugoBinaryTestHelper) InitModule(dir, modulePath string) error {
	h.t.Helper()

	// Change to directory and run hugo mod init
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		if chdirErr := os.Chdir(oldDir); chdirErr != nil {
			h.t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(dir); chdirErr != nil {
		return chdirErr
	}

	_, err = h.RunHugo("mod", "init", modulePath)
	return err
}

// GetModule gets a Hugo module (e.g., github.com/imfing/hextra)
func (h *HugoBinaryTestHelper) GetModule(dir, modulePath string) error {
	h.t.Helper()

	// Change to directory and run hugo mod get
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		if chdirErr := os.Chdir(oldDir); chdirErr != nil {
			h.t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(dir); chdirErr != nil {
		return chdirErr
	}

	_, err = h.RunHugo("mod", "get", modulePath)
	return err
}

// BuildSite builds the Hugo site
func (h *HugoBinaryTestHelper) BuildSite(dir string) error {
	h.t.Helper()

	// Change to directory and run hugo build
	oldDir, err := os.Getwd()
	if err != nil {
		return err
	}
	defer func() {
		if chdirErr := os.Chdir(oldDir); chdirErr != nil {
			h.t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(dir); chdirErr != nil {
		return chdirErr
	}

	_, err = h.RunHugo("--gc", "--minify")
	return err
}

// Cleanup removes temporary files and directories
func (h *HugoBinaryTestHelper) Cleanup() {
	h.t.Helper()
	if h.TempDir != "" {
		if err := os.RemoveAll(h.TempDir); err != nil {
			h.t.Logf("Failed to remove temp directory: %v", err)
		}
		h.t.Logf("Cleaned up Hugo binary temp directory: %s", h.TempDir)
	}
}

// SkipIfDownloadFails attempts to download Hugo and skips test if it fails
func (h *HugoBinaryTestHelper) SkipIfDownloadFails() {
	h.t.Helper()

	if err := h.DownloadAndExtract(); err != nil {
		h.t.Skipf("Skipping test because Hugo binary download failed: %v", err)
	}
}
