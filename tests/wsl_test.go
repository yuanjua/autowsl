package tests

import (
	"strings"
	"testing"

	"github.com/yuanjua/autowsl/internal/wsl"
)

// MockRunner is a test double for runner.Runner
type MockRunner struct {
	Outputs map[string]string // command -> stdout
	Errors  map[string]error  // command -> error
	Calls   []string          // track all commands called
}

func NewMockRunner() *MockRunner {
	return &MockRunner{
		Outputs: make(map[string]string),
		Errors:  make(map[string]error),
		Calls:   make([]string, 0),
	}
}

func (m *MockRunner) Run(name string, args ...string) (string, string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.Calls = append(m.Calls, cmd)

	if err, ok := m.Errors[cmd]; ok {
		return "", "", err
	}
	if output, ok := m.Outputs[cmd]; ok {
		return output, "", nil
	}
	return "", "", nil
}

func (m *MockRunner) RunWithInput(name string, stdin string, args ...string) (string, string, error) {
	return m.Run(name, args...)
}

func TestWSLListInstalledDistros(t *testing.T) {
	// Prepare fake WSL output
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  kali-linux             Stopped         1
`

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

	client := wsl.NewClient(mock)
	distros, err := client.ListInstalledDistros()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify command was called
	if len(mock.Calls) != 1 {
		t.Errorf("Expected 1 call, got %d", len(mock.Calls))
	}

	// Verify we got 3 distributions
	if len(distros) != 3 {
		t.Fatalf("Expected 3 distros, got %d", len(distros))
	}

	// Test first distro (default)
	if distros[0].Name != "Ubuntu-22.04" {
		t.Errorf("Expected name 'Ubuntu-22.04', got '%s'", distros[0].Name)
	}
	if distros[0].State != "Running" {
		t.Errorf("Expected state 'Running', got '%s'", distros[0].State)
	}
	if distros[0].Version != "2" {
		t.Errorf("Expected version '2', got '%s'", distros[0].Version)
	}
	if !distros[0].Default {
		t.Error("Expected first distro to be default")
	}

	// Test second distro (not default)
	if distros[1].Name != "Debian" {
		t.Errorf("Expected name 'Debian', got '%s'", distros[1].Name)
	}
	if distros[1].Default {
		t.Error("Expected second distro not to be default")
	}

	// Test third distro (WSL 1)
	if distros[2].Version != "1" {
		t.Errorf("Expected version '1', got '%s'", distros[2].Version)
	}
}

func TestWSLIsDistroInstalled(t *testing.T) {
	fakeOutput := `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
`

	mock := NewMockRunner()
	mock.Outputs["wsl.exe -l -v"] = fakeOutput

	client := wsl.NewClient(mock)

	tests := []struct {
		name     string
		distro   string
		expected bool
	}{
		{"existing distro", "Ubuntu-22.04", true},
		{"another existing distro", "Debian", true},
		{"non-existing distro", "Arch", false},
		{"case sensitive check", "ubuntu-22.04", false}, // Names are case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := client.IsDistroInstalled(tt.distro)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
			if exists != tt.expected {
				t.Errorf("IsDistroInstalled(%s) = %v, want %v", tt.distro, exists, tt.expected)
			}
		})
	}
}

func TestWSLCheckWSLInstalled(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockRunner)
		expectError bool
	}{
		{
			name: "WSL is installed",
			setupMock: func(m *MockRunner) {
				m.Outputs["wsl.exe --status"] = "some status output"
			},
			expectError: false,
		},
		{
			name: "WSL is not installed",
			setupMock: func(m *MockRunner) {
				m.Errors["wsl.exe --status"] = &mockError{"wsl not found"}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRunner()
			tt.setupMock(mock)

			client := wsl.NewClient(mock)
			err := client.CheckWSLInstalled()

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestWSLParseListEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCount int
		description   string
	}{
		{
			name: "UTF-8 BOM and null bytes",
			input: "\ufeff\x00  NAME\x00                   STATE\x00           VERSION\x00\n" +
				"* Ubuntu\x00                 Running\x00         2\x00\n",
			expectedCount: 1,
			description:   "Should handle UTF-8 BOM and null bytes",
		},
		{
			name:          "Empty output",
			input:         "",
			expectedCount: 0,
			description:   "Should handle empty output",
		},
		{
			name: "Only headers",
			input: `  NAME                   STATE           VERSION
----------------------------------------
`,
			expectedCount: 0,
			description:   "Should handle output with only headers",
		},
		{
			name: "Multiple distros with varying states",
			input: `  NAME                   STATE           VERSION
* Ubuntu-22.04           Running         2
  Debian                 Stopped         2
  Arch                   Installing      2
  Alpine                 Uninstalling    1
`,
			expectedCount: 4,
			description:   "Should parse multiple distros with different states",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockRunner()
			mock.Outputs["wsl.exe -l -v"] = tt.input

			client := wsl.NewClient(mock)
			distros, err := client.ListInstalledDistros()

			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if len(distros) != tt.expectedCount {
				t.Errorf("%s: expected %d distros, got %d", tt.description, tt.expectedCount, len(distros))
			}
		})
	}
}

// Helper type for mock errors
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
