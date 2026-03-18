package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// writeGlobalConfig creates ~/.calf/config.yaml in home with the given YAML content.
func writeGlobalConfig(t *testing.T, home, yamlContent string) {
	t.Helper()
	configDir := filepath.Join(home, ".calf")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
}

// setupConfigShow creates a fresh root command for "calf config show [extraArgs...]"
// in an isolated temp HOME. Returns the command, home dir, and captured stdout/stderr.
func setupConfigShow(t *testing.T, extraArgs ...string) (cmd *cobra.Command, home string, out, errOut *bytes.Buffer) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	out = &bytes.Buffer{}
	errOut = &bytes.Buffer{}
	cmd = newRootCmd("test")
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(append([]string{"config", "show"}, extraArgs...))
	return cmd, home, out, errOut
}

func TestConfigShow(t *testing.T) {
	t.Run("when config file is missing should output default values", func(t *testing.T) {
		// Arrange
		cmd, _, out, _ := setupConfigShow(t)

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "ghcr.io/cirruslabs/macos-sequoia-base:latest") {
			t.Errorf("expected default base image in output, got: %s", out.String())
		}
	})

	t.Run("when valid config file exists should output base image field", func(t *testing.T) {
		// Arrange
		cmd, home, out, _ := setupConfigShow(t)
		writeGlobalConfig(t, home, "version: 1\nisolation:\n  defaults:\n    vm:\n      base_image: ghcr.io/custom/image:v1\n")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "ghcr.io/custom/image:v1") {
			t.Errorf("expected custom base image in output, got: %s", out.String())
		}
	})

	t.Run("when valid config file exists should output cpu count field", func(t *testing.T) {
		// Arrange
		cmd, home, out, _ := setupConfigShow(t)
		writeGlobalConfig(t, home, "version: 1\nisolation:\n  defaults:\n    vm:\n      cpu: 8\n")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "8 cores") {
			t.Errorf("expected cpu count in output, got: %s", out.String())
		}
	})

	t.Run("when valid config file exists should output memory size field", func(t *testing.T) {
		// Arrange
		cmd, home, out, _ := setupConfigShow(t)
		writeGlobalConfig(t, home, "version: 1\nisolation:\n  defaults:\n    vm:\n      memory: 16384\n")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "16384 MB") {
			t.Errorf("expected memory size in output, got: %s", out.String())
		}
	})

	t.Run("when vm name flag provided and vm config exists should output vm-specific values", func(t *testing.T) {
		// Arrange
		cmd, home, out, _ := setupConfigShow(t, "--vm", "test-vm")
		writeGlobalConfig(t, home, "version: 1\nisolation:\n  defaults:\n    vm:\n      base_image: ghcr.io/global/image:v1\n")
		vmDir := filepath.Join(home, ".calf", "isolation", "vms", "test-vm")
		if err := os.MkdirAll(vmDir, 0755); err != nil {
			t.Fatal(err)
		}
		vmYAML := "base_image: ghcr.io/vm-specific/image:v2\n"
		if err := os.WriteFile(filepath.Join(vmDir, "vm.yaml"), []byte(vmYAML), 0644); err != nil {
			t.Fatal(err)
		}

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "ghcr.io/vm-specific/image:v2") {
			t.Errorf("expected vm-specific base image in output, got: %s", out.String())
		}
	})

	t.Run("when vm name flag provided and vm config missing should fall back to global config", func(t *testing.T) {
		// Arrange
		cmd, home, out, _ := setupConfigShow(t, "--vm", "nonexistent-vm")
		writeGlobalConfig(t, home, "version: 1\nisolation:\n  defaults:\n    vm:\n      base_image: ghcr.io/global/image:v1\n")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "ghcr.io/global/image:v1") {
			t.Errorf("expected global base image as fallback, got: %s", out.String())
		}
	})

	t.Run("when config file path is invalid should return error not exit process", func(t *testing.T) {
		// Arrange
		cmd, home, _, _ := setupConfigShow(t)
		writeGlobalConfig(t, home, "[invalid yaml")

		// Act
		err := cmd.Execute()

		// Assert
		if err == nil {
			t.Fatal("expected error when config file is invalid, got nil")
		}
	})
}
