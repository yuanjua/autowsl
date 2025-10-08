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

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

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

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

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

func TestListInstalledDistrosFallbackBasic(t *testing.T) {
	// Simulate failure of -l -v and success of -l
	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l"] = "Windows Subsystem for Linux Distributions:\nUbuntu-22.04 (Default)\nDebian\n"
	mock.Errors["wsl.exe -l -v"] = &mockError{"exit status 0xffffffff"}

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()
	if err != nil {
		t.Fatalf("Expected no error on fallback, got %v", err)
	}
	if len(distros) != 2 {
		t.Fatalf("Expected 2 distros, got %d", len(distros))
	}
	if distros[0].Name != "Ubuntu-22.04" || !distros[0].Default {
		t.Errorf("Expected Ubuntu-22.04 default, got %+v", distros[0])
	}
}

func TestListInstalledDistrosBenignFailure(t *testing.T) {
	// Simulate both commands failing with a benign pre-initialization error
	mock := NewMockRunner()
	mock.Errors["wsl.exe -l -v"] = &mockError{"exit status 0xffffffff"}
	mock.Errors["wsl.exe -l"] = &mockError{"exit status 0xffffffff"}

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()
	if err != nil {
		t.Fatalf("Expected benign failure to return empty slice without error, got %v", err)
	}
	if len(distros) != 0 {
		t.Fatalf("Expected 0 distros, got %d", len(distros))
	}
}
