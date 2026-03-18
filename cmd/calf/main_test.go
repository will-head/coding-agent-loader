package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// setupRootCmd creates a fresh root command configured for testing with
// captured stdout/stderr. Returns the command and output buffers.
func setupRootCmd(t *testing.T, args ...string) (cmd *cobra.Command, out, errOut *bytes.Buffer) {
	t.Helper()
	out = &bytes.Buffer{}
	errOut = &bytes.Buffer{}
	cmd = newRootCmd("test")
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs(args)
	return cmd, out, errOut
}

func TestRootCommand(t *testing.T) {
	t.Run("when no args provided should print usage information", func(t *testing.T) {
		// Arrange
		cmd, out, _ := setupRootCmd(t)

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "Usage:") {
			t.Errorf("expected usage information in output, got: %s", out.String())
		}
	})

	t.Run("when help flag provided should print help text", func(t *testing.T) {
		// Arrange
		cmd, out, _ := setupRootCmd(t, "--help")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "CALF") {
			t.Errorf("expected CALF in help output, got: %s", out.String())
		}
	})

	t.Run("when unknown subcommand provided should return error", func(t *testing.T) {
		// Arrange
		cmd, _, _ := setupRootCmd(t, "unknowncmd")

		// Act
		err := cmd.Execute()

		// Assert
		if err == nil {
			t.Fatal("expected error for unknown subcommand, got nil")
		}
	})
}

func TestConfigSubcommand(t *testing.T) {
	t.Run("when config subcommand provided should be recognized", func(t *testing.T) {
		// Arrange
		cmd, out, _ := setupRootCmd(t, "config")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected config subcommand to be recognized, got error: %v", err)
		}
		if !strings.Contains(out.String(), "config") {
			t.Errorf("expected config in output, got: %s", out.String())
		}
	})
}

func TestCacheSubcommand(t *testing.T) {
	t.Run("when cache subcommand provided should be recognized", func(t *testing.T) {
		// Arrange
		cmd, out, _ := setupRootCmd(t, "cache")

		// Act
		err := cmd.Execute()

		// Assert
		if err != nil {
			t.Fatalf("expected cache subcommand to be recognized, got error: %v", err)
		}
		if !strings.Contains(out.String(), "cache") {
			t.Errorf("expected cache in output, got: %s", out.String())
		}
	})
}
