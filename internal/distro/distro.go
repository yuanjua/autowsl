package distro

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed distros-winget.json
var distrosJSON []byte

// Distro represents a WSL distribution
type Distro struct {
	Group        string `json:"group"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	PackageID    string `json:"packageId,omitempty"` // Winget package ID (new method)
	URL          string `json:"url,omitempty"`       // Direct URL (legacy method)
	SHA256       string `json:"sha256,omitempty"`    // Optional checksum for verification
}

// DistroList represents the JSON structure
type DistroList struct {
	Distributions []Distro `json:"distributions"`
}

// GetAllDistros returns all available WSL distributions from embedded JSON
func GetAllDistros() []Distro {
	var distroList DistroList

	if err := json.Unmarshal(distrosJSON, &distroList); err != nil {
		// Fallback to empty list if JSON parsing fails
		fmt.Printf("Warning: Failed to parse distros.json: %v\n", err)
		return []Distro{}
	}

	return distroList.Distributions
}

// FindDistroByVersion finds a distribution by its version name
func FindDistroByVersion(version string) (*Distro, error) {
	distros := GetAllDistros()

	for _, d := range distros {
		if d.Version == version {
			return &d, nil
		}
	}

	return nil, fmt.Errorf("distribution '%s' not found", version)
}

// GetDistrosByGroup returns all distributions in a specific group
func GetDistrosByGroup(group string) []Distro {
	distros := GetAllDistros()
	var result []Distro

	for _, d := range distros {
		if d.Group == group {
			result = append(result, d)
		}
	}

	return result
}
