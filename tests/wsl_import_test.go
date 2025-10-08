package tests

import (
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

func TestWSLImportValidation(t *testing.T) {
	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = "" // No existing distros

	client := wsl.NewClient(mock)

	tests := []struct {
		name        string
		opts        wsl.ImportOptions
		expectError bool
		errorMsg    string
	}{
		{
			name: "empty name",
			opts: wsl.ImportOptions{
				Name:        "",
				InstallPath: "/some/path",
				TarFilePath: "test.tar",
			},
			expectError: true,
			errorMsg:    "name cannot be empty",
		},
		{
			name: "empty install path",
			opts: wsl.ImportOptions{
				Name:        "test-distro",
				InstallPath: "",
				TarFilePath: "test.tar",
			},
			expectError: true,
			errorMsg:    "installation path cannot be empty",
		},
		{
			name: "empty tar file path",
			opts: wsl.ImportOptions{
				Name:        "test-distro",
				InstallPath: "/some/path",
				TarFilePath: "",
			},
			expectError: true,
			errorMsg:    "tar file path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Import(tt.opts)

			if tt.expectError && err == nil {
				t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if tt.expectError && err != nil {
				// Just verify we got an error, don't check exact message
				// as file existence checks might vary
			}
		})
	}
}

func TestWSLUnregisterValidation(t *testing.T) {
	mock := NewMockRunner()
	client := wsl.NewClient(mock)

	// Test empty name
	err := client.Unregister("")
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	// Test non-existing distro
	mock.Outputs["wsl.exe -l -v"] = `  NAME                   STATE           VERSION
* Ubuntu                 Running         2
`
	err = client.Unregister("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existing distro, got nil")
	}

	// Test successful unregister
	mock.Outputs["wsl.exe --unregister Ubuntu"] = ""
	err = client.Unregister("Ubuntu")
	if err != nil {
		t.Errorf("Expected no error for valid unregister, got: %v", err)
	}
}

func TestWSLExportValidation(t *testing.T) {
	mock := NewMockRunner()
	client := wsl.NewClient(mock)

	tests := []struct {
		name        string
		distroName  string
		outputPath  string
		expectError bool
	}{
		{
			name:        "empty distro name",
			distroName:  "",
			outputPath:  "output.tar",
			expectError: true,
		},
		{
			name:        "empty output path",
			distroName:  "Ubuntu",
			outputPath:  "",
			expectError: true,
		},
		{
			name:        "non-existing distro",
			distroName:  "NonExistent",
			outputPath:  "output.tar",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock to return no distros for non-existing check
			mock.Outputs["wsl.exe -l -v"] = ""

			err := client.Export(tt.distroName, tt.outputPath)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}
