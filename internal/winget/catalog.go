package winget

import (
	"github.com/yuanjua/autowsl/internal/distro"
)

// WingetDistro represents a WSL distribution available via winget
type WingetDistro struct {
	Name         string // Display name
	Version      string // Version string
	PackageID    string // Winget package identifier
	Group        string // Distribution family (Ubuntu, Debian, etc.)
	Architecture string // Supported architectures
}

// GetWingetDistros returns the list of WSL distributions available via winget
// This reads from the distro catalog
func GetWingetDistros() []WingetDistro {
	distros := distro.GetAllDistros()
	result := make([]WingetDistro, 0, len(distros))

	for _, d := range distros {
		if d.PackageID != "" {
			result = append(result, WingetDistro{
				Name:         d.Version,
				Version:      d.Version,
				PackageID:    d.PackageID,
				Group:        d.Group,
				Architecture: d.Architecture,
			})
		}
	}

	return result
}

// FindWingetDistroByVersion finds a distribution by its version string
func FindWingetDistroByVersion(version string) (*WingetDistro, error) {
	d, err := distro.FindDistroByVersion(version)
	if err != nil {
		return nil, err
	}

	if d.PackageID == "" {
		return nil, nil // No winget package ID
	}

	return &WingetDistro{
		Name:         d.Version,
		Version:      d.Version,
		PackageID:    d.PackageID,
		Group:        d.Group,
		Architecture: d.Architecture,
	}, nil
}

// FindWingetDistroByPackageID finds a distribution by its winget package ID
func FindWingetDistroByPackageID(packageID string) (*WingetDistro, error) {
	distros := distro.GetAllDistros()
	for _, d := range distros {
		if d.PackageID == packageID {
			return &WingetDistro{
				Name:         d.Version,
				Version:      d.Version,
				PackageID:    d.PackageID,
				Group:        d.Group,
				Architecture: d.Architecture,
			}, nil
		}
	}
	return nil, nil // Not found
}
