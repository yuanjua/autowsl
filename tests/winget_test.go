package tests

import (
	"fmt"
	"testing"

	"github.com/yuanjua/autowsl/internal/winget"
)

// TestWingetDistrosCatalog verifies the catalog structure
func TestWingetDistrosCatalog(t *testing.T) {
	distros := winget.GetWingetDistros()

	if len(distros) == 0 {
		t.Fatal("Catalog should not be empty")
	}

	// Verify each distro has required fields
	for _, d := range distros {
		if d.Version == "" {
			t.Errorf("Distro missing Version: %+v", d)
		}
		if d.PackageID == "" {
			t.Errorf("Distro missing PackageID: %+v", d)
		}
		if d.Group == "" {
			t.Errorf("Distro missing Group: %+v", d)
		}
	}

	t.Logf("Catalog contains %d distributions", len(distros))
}

// TestFindWingetDistroByVersion tests version lookup
func TestFindWingetDistroByVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"Ubuntu 22.04 LTS", false},
		{"Ubuntu 24.04 LTS", false},
		{"Debian", false},
		{"NonExistent", true}, // Should return error for non-existent
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			distro, err := winget.FindWingetDistroByVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindWingetDistroByVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && distro == nil {
				t.Errorf("Expected distro for %s, got nil", tt.version)
			}
		})
	}
}

// TestIsWSLPackage tests the WSL package detection heuristic
func TestIsWSLPackage(t *testing.T) {
	tests := []struct {
		packageID string
		want      bool
	}{
		{"Canonical.Ubuntu.2204", true},
		{"Debian.Debian", true},
		{"KaliLinux.KaliLinux", true},
		{"openSUSE.Tumbleweed", true},
		{"Microsoft.VisualStudioCode", false},
		{"Google.Chrome", false},
		{"SomeVendor.WSLDistro", true}, // Contains "WSL"
	}

	for _, tt := range tests {
		t.Run(tt.packageID, func(t *testing.T) {
			if got := winget.IsWSLPackage(tt.packageID); got != tt.want {
				t.Errorf("IsWSLPackage(%s) = %v, want %v", tt.packageID, got, tt.want)
			}
		})
	}
}

// ExampleManager demonstrates basic usage
func ExampleManager() {
	// Create download manager
	mgr := winget.NewManager("./.autowsl_tmp")

	// Check if wingetcreate is available
	if !mgr.IsWingetAvailable() {
		fmt.Println("Wingetcreate is not available")
		return
	}

	// List available distributions
	distros := mgr.GetCatalog()
	fmt.Printf("Available distributions: %d\n", len(distros))

	// Download a distribution (commented out to avoid actual download in example)
	// opts := winget.DownloadOptions{Version: "Ubuntu 22.04 LTS"}
	// filePath, _ := mgr.Download(opts)
	// fmt.Printf("Downloaded to: %s\n", filePath)
}

// ExampleDownloadOptions_catalog demonstrates downloading from catalog
func ExampleDownloadOptions_catalog() {
	mgr := winget.NewManager("./.autowsl_tmp")

	opts := winget.DownloadOptions{
		Version: "Ubuntu 22.04 LTS",
	}

	filePath, err := mgr.Download(opts)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded to: %s\n", filePath)
}

// ExampleDownloadOptions_custom demonstrates downloading with custom package ID
func ExampleDownloadOptions_custom() {
	mgr := winget.NewManager("./.autowsl_tmp")

	opts := winget.DownloadOptions{
		PackageID:         "Canonical.Ubuntu.2204",
		ValidatePackageID: true,
	}

	filePath, err := mgr.Download(opts)
	if err != nil {
		fmt.Printf("Download failed: %v\n", err)
		return
	}

	fmt.Printf("Downloaded to: %s\n", filePath)
}
