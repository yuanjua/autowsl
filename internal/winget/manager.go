package winget

import (
	"fmt"
)

// Manager handles WSL distribution downloads using winget
type Manager struct {
	downloader *WingetDownloader
	tempDir    string
}

// NewManager creates a new download manager
func NewManager(tempDir string) *Manager {
	return &Manager{
		downloader: NewWingetDownloader(tempDir),
		tempDir:    tempDir,
	}
}

// DownloadOptions contains options for downloading a distribution
type DownloadOptions struct {
	// Either specify a known version or a custom package ID
	Version   string // e.g., "Ubuntu 22.04 LTS"
	PackageID string // e.g., "Canonical.Ubuntu.2204"

	// If true, validate package ID before downloading
	ValidatePackageID bool
}

// Download downloads a WSL distribution
// Returns the path to the downloaded file
func (m *Manager) Download(opts DownloadOptions) (string, error) {
	var packageID string

	// Determine package ID
	if opts.PackageID != "" {
		// User provided a custom package ID
		packageID = opts.PackageID

		// Optionally validate it
		if opts.ValidatePackageID {
			fmt.Printf("Validating package ID: %s\n", packageID)
			valid, err := ValidatePackageID(packageID)
			if err != nil {
				return "", fmt.Errorf("failed to validate package ID: %w", err)
			}
			if !valid {
				return "", fmt.Errorf("package ID '%s' not found in winget", packageID)
			}

			// Warn if it doesn't look like a WSL package
			if !IsWSLPackage(packageID) {
				fmt.Printf("Warning: Package '%s' may not be a WSL distribution\n", packageID)
			}
		}
	} else if opts.Version != "" {
		// Look up version in catalog
		distro, err := FindWingetDistroByVersion(opts.Version)
		if err != nil {
			return "", err
		}

		if distro == nil {
			return "", fmt.Errorf("distribution '%s' not found in catalog. Use --package-id to specify a custom winget package", opts.Version)
		}

		packageID = distro.PackageID
		fmt.Printf("Found in catalog: %s (%s)\n", distro.Name, distro.PackageID)
	} else {
		return "", fmt.Errorf("either Version or PackageID must be specified")
	}

	// Download using winget
	downloadedFile, err := m.downloader.Download(packageID)
	if err != nil {
		return "", err
	}

	return downloadedFile, nil
}

// GetCatalog returns the list of known distributions
func (m *Manager) GetCatalog() []WingetDistro {
	return GetWingetDistros()
}

// IsWingetAvailable checks if winget is available
func (m *Manager) IsWingetAvailable() bool {
	return m.downloader.IsWingetAvailable()
}

// GetWingetVersion returns the winget version
func (m *Manager) GetWingetVersion() (string, error) {
	return m.downloader.GetWingetVersion()
}

// CleanupDownloadDir removes the temporary download directory
func (m *Manager) CleanupDownloadDir() error {
	// Don't cleanup if user might want to keep files
	// This should be called explicitly
	fmt.Printf("Download directory: %s\n", m.tempDir)
	fmt.Println("Files are kept for your use. Delete manually if needed.")
	return nil
}
