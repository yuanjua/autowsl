package tests

import (
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

func TestListInstalledDistros(t *testing.T) {
	// Fake wsl.exe output
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  kali-linux             Stopped         2
`

	mock := &MockRunner{
		Outputs: map[string]string{
			"wsl.exe -l -v": fakeOutput,
		},
	}

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(distros) != 3 {
		t.Fatalf("Expected 3 distros, got %d", len(distros))
	}

	// Check first distro
	if distros[0].Name != "Ubuntu-22.04" {
		t.Errorf("Expected 'Ubuntu-22.04', got '%s'", distros[0].Name)
	}
	if distros[0].State != "Running" {
		t.Errorf("Expected 'Running', got '%s'", distros[0].State)
	}
	if !distros[0].Default {
		t.Error("Expected first distro to be default")
	}

	// Check second distro
	if distros[1].Name != "Debian" {
		t.Errorf("Expected 'Debian', got '%s'", distros[1].Name)
	}
	if distros[1].Default {
		t.Error("Expected second distro not to be default")
	}
}

func TestIsDistroInstalled(t *testing.T) {
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
`

	mock := &MockRunner{
		Outputs: map[string]string{
			"wsl.exe -l -v": fakeOutput,
		},
	}

	client := wsl.NewClient(mock)

	// Test existing distro
	exists, err := client.IsDistroInstalled("Ubuntu-22.04")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected Ubuntu-22.04 to exist")
	}

	// Test non-existing distro
	exists, err = client.IsDistroInstalled("NonExistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected NonExistent to not exist")
	}
}
