// Package isolation provides VM isolation and management for CALF.
package isolation

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

// mockCommandRunner is a test helper that simulates command execution
type mockCommandRunner struct {
	commands [][]string
	outputs  map[string]string
	errors   map[string]error
}

func newMockCommandRunner() *mockCommandRunner {
	return &mockCommandRunner{
		commands: make([][]string, 0),
		outputs:  make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *mockCommandRunner) addOutput(cmdKey string, output string) {
	m.outputs[cmdKey] = output
}

func (m *mockCommandRunner) addError(cmdKey string, err error) {
	m.errors[cmdKey] = err
}

func (m *mockCommandRunner) runCommand(name string, args ...string) (string, error) {
	cmdArgs := append([]string{name}, args...)
	m.commands = append(m.commands, cmdArgs)

	cmdKey := strings.Join(args, " ")
	if err, ok := m.errors[cmdKey]; ok {
		return "", err
	}
	if output, ok := m.outputs[cmdKey]; ok {
		return output, nil
	}
	return "", nil
}

// createTestClient creates a TartClient configured for testing.
// Extra options override the defaults (e.g. WithPollTimeout for shorter timeouts).
func createTestClient(mock *mockCommandRunner, extra ...TartClientOption) *TartClient {
	return NewTartClient(append([]TartClientOption{
		WithTartPath("/usr/local/bin/tart"),
		WithPollInterval(10 * time.Millisecond),
		WithPollTimeout(100 * time.Millisecond),
		WithRunCommand(func(args ...string) (string, error) {
			return mock.runCommand("tart", args...)
		}),
	}, extra...)...)
}

func TestVMStateString(t *testing.T) {
	t.Run("when state is running should return string running", func(t *testing.T) {
		// Arrange
		state := StateRunning

		// Act
		got := string(state)

		// Assert
		if got != "running" {
			t.Errorf("VMState string = %q, want %q", got, "running")
		}
	})

	t.Run("when state is stopped should return string stopped", func(t *testing.T) {
		// Arrange
		state := StateStopped

		// Act
		got := string(state)

		// Assert
		if got != "stopped" {
			t.Errorf("VMState string = %q, want %q", got, "stopped")
		}
	})

	t.Run("when state is not found should return string not_found", func(t *testing.T) {
		// Arrange
		state := StateNotFound

		// Act
		got := string(state)

		// Assert
		if got != "not_found" {
			t.Errorf("VMState string = %q, want %q", got, "not_found")
		}
	})
}

func TestClone(t *testing.T) {
	t.Run("when clone succeeds should execute tart clone command with correct args", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("clone test-image test-vm", "")
		client := createTestClient(mock)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err != nil {
			t.Errorf("Clone() unexpected error = %v", err)
		}
		if len(mock.commands) != 1 {
			t.Errorf("Expected 1 command, got %d", len(mock.commands))
		}
		expected := []string{"tart", "clone", "test-image", "test-vm"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Clone() command = %v, want %v", mock.commands[0], expected)
		}
	})

	t.Run("when clone fails should return wrapped error", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addError("clone test-image test-vm", fmt.Errorf("clone failed"))
		client := createTestClient(mock)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err == nil {
			t.Error("Clone() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to clone VM") {
			t.Errorf("Clone() error should contain context, got: %v", err)
		}
	})
}

func TestSet(t *testing.T) {
	t.Run("when all params provided should include cpu memory and disk size flags", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("set test-vm --cpu=4 --memory=8192 --disk-size=80", "")
		client := createTestClient(mock)

		// Act
		err := client.Set("test-vm", 4, 8192, "80")

		// Assert
		if err != nil {
			t.Errorf("Set() unexpected error = %v", err)
		}
		expected := []string{"tart", "set", "test-vm", "--cpu=4", "--memory=8192", "--disk-size=80"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Set() command = %v, want %v", mock.commands[0], expected)
		}
	})

	t.Run("when only cpu provided should include only cpu flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("set test-vm --cpu=4", "")
		client := createTestClient(mock)

		// Act
		err := client.Set("test-vm", 4, 0, "")

		// Assert
		if err != nil {
			t.Errorf("Set() unexpected error = %v", err)
		}
		expected := []string{"tart", "set", "test-vm", "--cpu=4"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Set() command = %v, want %v", mock.commands[0], expected)
		}
	})
}

func TestStop(t *testing.T) {
	t.Run("when force is false should execute stop without timeout flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("stop test-vm", "")
		client := createTestClient(mock)

		// Act
		err := client.Stop("test-vm", false)

		// Assert
		if err != nil {
			t.Errorf("Stop() unexpected error = %v", err)
		}
		expected := []string{"tart", "stop", "test-vm"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Stop() command = %v, want %v", mock.commands[0], expected)
		}
	})

	t.Run("when force is true should pass timeout zero flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("stop test-vm --timeout=0", "")
		client := createTestClient(mock)

		// Act
		err := client.Stop("test-vm", true)

		// Assert
		if err != nil {
			t.Errorf("Stop() unexpected error = %v", err)
		}
		expected := []string{"tart", "stop", "test-vm", "--timeout=0"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Stop() command = %v, want %v", mock.commands[0], expected)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("when vm exists should execute tart delete with vm name", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("delete test-vm", "")
		client := createTestClient(mock)

		// Act
		err := client.Delete("test-vm")

		// Assert
		if err != nil {
			t.Errorf("Delete() unexpected error = %v", err)
		}
		expected := []string{"tart", "delete", "test-vm"}
		if !slices.Equal(mock.commands[0], expected) {
			t.Errorf("Delete() command = %v, want %v", mock.commands[0], expected)
		}
	})
}

func TestList(t *testing.T) {
	t.Run("when tart returns valid json should parse vm list", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		jsonOutput := `[
		{"name":"calf-dev","state":"running","size":10.5},
		{"name":"calf-clean","state":"stopped","size":8.2}
	]`
		mock.addOutput("list --format json", jsonOutput)
		client := createTestClient(mock)

		// Act
		vms, err := client.List()

		// Assert
		if err != nil {
			t.Errorf("List() unexpected error = %v", err)
		}
		if len(vms) != 2 {
			t.Errorf("List() returned %d VMs, want 2", len(vms))
		}
		if vms[0].Name != "calf-dev" || vms[0].State != StateRunning {
			t.Errorf("List() first VM = %+v, want calf-dev running", vms[0])
		}
		if vms[1].Name != "calf-clean" || vms[1].State != StateStopped {
			t.Errorf("List() second VM = %+v, want calf-clean stopped", vms[1])
		}
	})

	t.Run("when tart returns invalid json should return parse error", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", "invalid json")
		client := createTestClient(mock)

		// Act
		_, err := client.List()

		// Assert
		if err == nil {
			t.Error("List() expected error for invalid JSON, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse VM list JSON") {
			t.Errorf("List() error should indicate JSON parse failure, got: %v", err)
		}
	})
}

func TestIP(t *testing.T) {
	t.Run("when vm acquires ip should return ip address", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("ip test-vm", "192.168.64.10\n")
		client := createTestClient(mock)

		// Act
		ip, err := client.IP("test-vm", 0)

		// Assert
		if err != nil {
			t.Errorf("IP() unexpected error = %v", err)
		}
		if ip != "192.168.64.10" {
			t.Errorf("IP() = %v, want 192.168.64.10", ip)
		}
	})

	t.Run("when vm does not acquire ip within timeout should return error", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addError("ip test-vm", fmt.Errorf("vm not ready"))
		client := createTestClient(mock, WithPollTimeout(50*time.Millisecond))

		// Act
		_, err := client.IP("test-vm", 0)

		// Assert
		if err == nil {
			t.Error("IP() expected timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "did not acquire an IP") {
			t.Errorf("IP() error should indicate timeout, got: %v", err)
		}
	})
}

func TestGet(t *testing.T) {
	t.Run("when vm exists in list should return vm info", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		jsonOutput := `[
		{"name":"calf-dev","state":"running","size":10.5},
		{"name":"test-vm","state":"stopped","size":8.2}
	]`
		mock.addOutput("list --format json", jsonOutput)
		client := createTestClient(mock)

		// Act
		vm, err := client.Get("test-vm")

		// Assert
		if err != nil {
			t.Errorf("Get() unexpected error = %v", err)
		}
		if vm.Name != "test-vm" || vm.State != StateStopped {
			t.Errorf("Get() = %+v, want test-vm stopped", vm)
		}
	})

	t.Run("when vm does not exist in list should return not found error", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		jsonOutput := `[{"name":"calf-dev","state":"running"}]`
		mock.addOutput("list --format json", jsonOutput)
		client := createTestClient(mock)

		// Act
		_, err := client.Get("nonexistent")

		// Assert
		if err == nil {
			t.Error("Get() expected error for nonexistent VM, got nil")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Get() error should indicate not found, got: %v", err)
		}
	})
}

func TestIsRunning(t *testing.T) {
	t.Run("when vm is running should return true", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[{"name":"test-vm","state":"running"}]`)
		client := createTestClient(mock)

		// Act
		got := client.IsRunning("test-vm")

		// Assert
		if !got {
			t.Errorf("IsRunning() = false, want true")
		}
	})

	t.Run("when vm is stopped should return false", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[{"name":"test-vm","state":"stopped"}]`)
		client := createTestClient(mock)

		// Act
		got := client.IsRunning("test-vm")

		// Assert
		if got {
			t.Errorf("IsRunning() = true, want false")
		}
	})

	t.Run("when vm does not exist should return false", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[]`)
		client := createTestClient(mock)

		// Act
		got := client.IsRunning("test-vm")

		// Assert
		if got {
			t.Errorf("IsRunning() = true, want false")
		}
	})
}

func TestExists(t *testing.T) {
	t.Run("when vm exists in list should return true", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[{"name":"test-vm","state":"running"}]`)
		client := createTestClient(mock)

		// Act
		got := client.Exists("test-vm")

		// Assert
		if !got {
			t.Errorf("Exists() = false, want true")
		}
	})

	t.Run("when vm does not exist in list should return false", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[]`)
		client := createTestClient(mock)

		// Act
		got := client.Exists("test-vm")

		// Assert
		if got {
			t.Errorf("Exists() = true, want false")
		}
	})
}

func TestGetState(t *testing.T) {
	t.Run("when vm is running should return running state", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[{"name":"test-vm","state":"running"}]`)
		client := createTestClient(mock)

		// Act
		got := client.GetState("test-vm")

		// Assert
		if got != StateRunning {
			t.Errorf("GetState() = %v, want StateRunning", got)
		}
	})

	t.Run("when vm is stopped should return stopped state", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[{"name":"test-vm","state":"stopped"}]`)
		client := createTestClient(mock)

		// Act
		got := client.GetState("test-vm")

		// Assert
		if got != StateStopped {
			t.Errorf("GetState() = %v, want StateStopped", got)
		}
	})

	t.Run("when vm does not exist should return not found state", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("list --format json", `[]`)
		client := createTestClient(mock)

		// Act
		got := client.GetState("test-vm")

		// Assert
		if got != StateNotFound {
			t.Errorf("GetState() = %v, want StateNotFound", got)
		}
	})
}

func TestRun(t *testing.T) {
	t.Run("when called with headless true should pass --headless flag to tart run", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.Run("test-vm", true, false, nil)

		// Assert
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Run() should have executed a command")
		}
		if !slices.Contains(mock.commands[0], "--headless") {
			t.Errorf("Run() command %v should contain --headless", mock.commands[0])
		}
	})

	t.Run("when called with headless false should not pass --headless flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.Run("test-vm", false, false, nil)

		// Assert
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Run() should have executed a command")
		}
		if slices.Contains(mock.commands[0], "--headless") {
			t.Errorf("Run() command %v should not contain --headless", mock.commands[0])
		}
	})

	t.Run("when called with vnc true should pass --vnc-experimental flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.Run("test-vm", false, true, nil)

		// Assert
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Run() should have executed a command")
		}
		if !slices.Contains(mock.commands[0], "--vnc-experimental") {
			t.Errorf("Run() command %v should contain --vnc-experimental", mock.commands[0])
		}
	})

	t.Run("when called with vnc false should not pass --vnc-experimental flag", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.Run("test-vm", false, false, nil)

		// Assert
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Run() should have executed a command")
		}
		if slices.Contains(mock.commands[0], "--vnc-experimental") {
			t.Errorf("Run() command %v should not contain --vnc-experimental", mock.commands[0])
		}
	})

	t.Run("when called should pass vm name as argument", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.Run("my-vm", false, false, nil)

		// Assert
		if err != nil {
			t.Errorf("Run() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Run() should have executed a command")
		}
		if !slices.Contains(mock.commands[0], "my-vm") {
			t.Errorf("Run() command %v should contain vm name 'my-vm'", mock.commands[0])
		}
	})
}

func TestRunWithCacheDirs(t *testing.T) {
	t.Run("when called with cache dirs should include --dir flag for each directory", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)
		cacheDirs := []string{"calf-cache:/path/to/cache", "npm-cache:/path/to/npm"}

		// Act
		err := client.RunWithCacheDirs("test-vm", false, false, nil, cacheDirs)

		// Assert
		if err != nil {
			t.Errorf("RunWithCacheDirs() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("RunWithCacheDirs() should have executed a command")
		}
		if !slices.Contains(mock.commands[0], "--dir=calf-cache:/path/to/cache") {
			t.Errorf("RunWithCacheDirs() command %v should contain --dir=calf-cache:/path/to/cache", mock.commands[0])
		}
		if !slices.Contains(mock.commands[0], "--dir=npm-cache:/path/to/npm") {
			t.Errorf("RunWithCacheDirs() command %v should contain --dir=npm-cache:/path/to/npm", mock.commands[0])
		}
	})

	t.Run("when called should always include the cache sharing directory", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.RunWithCacheDirs("test-vm", false, false, nil, nil)

		// Assert
		if err != nil {
			t.Errorf("RunWithCacheDirs() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("RunWithCacheDirs() should have executed a command")
		}
		expectedDir := fmt.Sprintf("--dir=%s", cacheDirMount)
		if !slices.Contains(mock.commands[0], expectedDir) {
			t.Errorf("RunWithCacheDirs() command %v should contain %s", mock.commands[0], expectedDir)
		}
	})

	t.Run("when called with empty cache dirs should still include cache sharing directory", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.RunWithCacheDirs("test-vm", false, false, nil, []string{})

		// Assert
		if err != nil {
			t.Errorf("RunWithCacheDirs() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("RunWithCacheDirs() should have executed a command")
		}
		expectedDir := fmt.Sprintf("--dir=%s", cacheDirMount)
		if !slices.Contains(mock.commands[0], expectedDir) {
			t.Errorf("RunWithCacheDirs() command %v should contain %s", mock.commands[0], expectedDir)
		}
	})

	t.Run("when called should pass vm name as argument", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		client := createTestClient(mock)

		// Act
		err := client.RunWithCacheDirs("my-vm", false, false, nil, nil)

		// Assert
		if err != nil {
			t.Errorf("RunWithCacheDirs() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("RunWithCacheDirs() should have executed a command")
		}
		if !slices.Contains(mock.commands[0], "my-vm") {
			t.Errorf("RunWithCacheDirs() command %v should contain vm name 'my-vm'", mock.commands[0])
		}
	})
}

func TestCloneWhenTartIsInstalled(t *testing.T) {
	t.Run("when tart is installed should dispatch clone command", func(t *testing.T) {
		// Arrange
		mock := newMockCommandRunner()
		mock.addOutput("clone test-image test-vm", "")
		client := createTestClient(mock)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err != nil {
			t.Errorf("Clone() unexpected error = %v", err)
		}
		if len(mock.commands) == 0 {
			t.Fatal("Clone() should have dispatched a command")
		}
		if !slices.Contains(mock.commands[0], "clone") {
			t.Errorf("Clone() command %v should contain 'clone'", mock.commands[0])
		}
	})
}

func TestCloneWhenTartIsNotInstalledAndUserDeclines(t *testing.T) {
	t.Run("when tart is not installed and user declines should return cancelled error", func(t *testing.T) {
		// Arrange
		client := NewTartClient(
			WithLookPath(func(file string) (string, error) {
				if file == "brew" {
					return "/usr/local/bin/brew", nil
				}
				return "", fmt.Errorf("not found")
			}),
			WithStdinReader(strings.NewReader("n\n")),
		)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err == nil {
			t.Error("Clone() expected error when user declines, got nil")
		}
		if !strings.Contains(err.Error(), "cancelled") {
			t.Errorf("Clone() error should indicate cancellation, got: %v", err)
		}
	})
}

func TestCloneWhenTartIsNotInstalledAndUserConfirmsAndBrewSucceeds(t *testing.T) {
	t.Run("when tart is not installed and user confirms and brew succeeds should clone successfully", func(t *testing.T) {
		// Arrange
		lookPathCalls := 0
		mock := newMockCommandRunner()
		mock.addOutput("clone test-image test-vm", "")
		client := NewTartClient(
			WithLookPath(func(file string) (string, error) {
				if file == "brew" {
					return "/usr/local/bin/brew", nil
				}
				if file == "tart" {
					lookPathCalls++
					if lookPathCalls > 1 {
						return "/usr/local/bin/tart", nil
					}
				}
				return "", fmt.Errorf("not found")
			}),
			WithStdinReader(strings.NewReader("y\n")),
			WithBrewRunner(func(args ...string) (string, error) {
				return "", nil
			}),
			WithRunCommand(func(args ...string) (string, error) {
				return mock.runCommand("tart", args...)
			}),
		)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err != nil {
			t.Errorf("Clone() unexpected error = %v", err)
		}
	})
}

func TestCloneWhenTartIsNotInstalledAndBrewFails(t *testing.T) {
	t.Run("when tart is not installed and brew fails should return install error", func(t *testing.T) {
		// Arrange
		client := NewTartClient(
			WithLookPath(func(file string) (string, error) {
				if file == "brew" {
					return "/usr/local/bin/brew", nil
				}
				return "", fmt.Errorf("not found")
			}),
			WithStdinReader(strings.NewReader("y\n")),
			WithBrewRunner(func(args ...string) (string, error) {
				return "", fmt.Errorf("brew install failed")
			}),
		)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err == nil {
			t.Error("Clone() expected error when brew fails, got nil")
		}
		if !strings.Contains(err.Error(), "failed to install") {
			t.Errorf("Clone() error should indicate install failure, got: %v", err)
		}
	})
}

func TestCloneWhenBrewIsNotAvailableAndTartNotFound(t *testing.T) {
	t.Run("when neither tart nor brew is available should return error without prompting", func(t *testing.T) {
		// Arrange
		client := NewTartClient(
			WithLookPath(func(file string) (string, error) {
				return "", fmt.Errorf("not found")
			}),
			WithStdinReader(strings.NewReader("")),
		)

		// Act
		err := client.Clone("test-image", "test-vm")

		// Assert
		if err == nil {
			t.Error("Clone() expected error when neither tart nor brew is found, got nil")
		}
	})
}

