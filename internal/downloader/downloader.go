package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuanjua/autowsl/internal/checksum"
	"github.com/yuanjua/autowsl/internal/distro"
)

// Downloader handles downloading WSL distributions
type Downloader struct {
	client         *http.Client
	VerifyChecksum bool // Whether to verify checksums (default: warn if mismatch)
}

// New creates a new Downloader instance
func New() *Downloader {
	return &Downloader{
		client:         &http.Client{},
		VerifyChecksum: false, // Default to warn-only mode
	}
}

// Download downloads a distribution to the current directory
func (d *Downloader) Download(dist distro.Distro) error {
	// Get the filename from the URL
	filename := d.getFilename(dist.URL)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	filepath := filepath.Join(cwd, filename)

	fmt.Printf("Downloading to: %s\n", filepath)

	return d.downloadToFile(dist.URL, filepath)
}

// DownloadToDir downloads a distribution to a specific directory and returns the file path
func (d *Downloader) DownloadToDir(dist distro.Distro, dir string) (string, error) {
	// Get the filename from the URL
	filename := d.getFilename(dist.URL)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	filepath := filepath.Join(dir, filename)

	// Always download fresh - remove any existing file first
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}

	if err := d.downloadToFile(dist.URL, filepath); err != nil {
		return "", err
	}

	// Verify checksum if provided
	if dist.SHA256 != "" {
		fmt.Println("Verifying checksum...")
		if err := checksum.VerifyFile(filepath, dist.SHA256); err != nil {
			if d.VerifyChecksum {
				// Strict mode: fail on mismatch
				os.Remove(filepath)
				return "", fmt.Errorf("checksum verification failed: %w", err)
			} else {
				// Warn mode: continue but alert user
				fmt.Printf("Warning: %v\n", err)
				fmt.Println("Continuing anyway (use --verify-checksum to enforce)")
			}
		} else {
			fmt.Println("Checksum verified successfully")
		}
	} else {
		fmt.Println("Warning: No checksum available for this distribution")
	}

	return filepath, nil
}

// downloadToFile downloads from URL to a specific file path
func (d *Downloader) downloadToFile(url, filepath string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", filepath, err)
	}
	defer out.Close()

	// Send GET request
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download from '%s': %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download from '%s': HTTP %s", url, resp.Status)
	}

	// Get the total size for progress
	totalSize := resp.ContentLength

	// Create progress writer
	counter := &ProgressWriter{
		Total:  totalSize,
		Writer: out,
	}

	// Copy the data with progress
	_, err = io.Copy(counter, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Println() // New line after progress

	return nil
}

// getFilename extracts filename from URL
func (d *Downloader) getFilename(url string) string {
	// Handle special cases for aka.ms URLs
	if strings.Contains(url, "aka.ms") {
		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]

		// Create a reasonable filename
		return fmt.Sprintf("%s.appxbundle", name)
	}

	// For direct URLs, extract the filename
	parts := strings.Split(url, "/")
	filename := parts[len(parts)-1]

	// Ensure it has a proper extension
	if !strings.Contains(filename, ".") {
		filename += ".appxbundle"
	}

	return filename
}

// ProgressWriter tracks download progress
type ProgressWriter struct {
	Total        int64
	Downloaded   int64
	Writer       io.Writer
	LastPrint    int64
	PrintEveryMB int64
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.Downloaded += int64(n)

	// Print progress every MB or at the end
	if pw.PrintEveryMB == 0 {
		pw.PrintEveryMB = 1024 * 1024 // 1 MB
	}

	if pw.Downloaded-pw.LastPrint >= pw.PrintEveryMB || pw.Downloaded == pw.Total {
		pw.printProgress()
		pw.LastPrint = pw.Downloaded
	}

	return n, nil
}

func (pw *ProgressWriter) printProgress() {
	if pw.Total > 0 {
		percentage := float64(pw.Downloaded) / float64(pw.Total) * 100
		downloadedMB := float64(pw.Downloaded) / 1024 / 1024
		totalMB := float64(pw.Total) / 1024 / 1024
		fmt.Printf("\rProgress: %.2f%% (%.2f MB / %.2f MB)", percentage, downloadedMB, totalMB)
	} else {
		downloadedMB := float64(pw.Downloaded) / 1024 / 1024
		fmt.Printf("\rDownloaded: %.2f MB", downloadedMB)
	}
}
