package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/will-head/coding-agent-launcher/internal/config"
)

// newConfigCmd constructs the config cobra command with all subcommands wired.
func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CALF configuration",
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Display effective configuration",
		RunE:  runConfigShow,
	}
	configShowCmd.Flags().StringP("vm", "v", "", "VM name to show config for")
	configCmd.AddCommand(configShowCmd)

	return configCmd
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	vmName, err := cmd.Flags().GetString("vm")
	if err != nil {
		return fmt.Errorf("getting vm flag: %w", err)
	}

	globalConfigPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return fmt.Errorf("getting default config path: %w", err)
	}

	var vmConfigPath string
	if vmName != "" {
		vmConfigPath, err = config.GetVMConfigPath(vmName)
		if err != nil {
			return fmt.Errorf("getting VM config path: %w", err)
		}
	}

	cfg, err := config.LoadConfig(globalConfigPath, vmConfigPath)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	out := cmd.OutOrStdout()

	fmt.Fprintln(out, "CALF Configuration")
	fmt.Fprintln(out, "=================")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "VM Defaults:")
	fmt.Fprintf(out, "  CPU: %d cores\n", cfg.Isolation.Defaults.VM.CPU)
	fmt.Fprintf(out, "  Memory: %d MB\n", cfg.Isolation.Defaults.VM.Memory)
	fmt.Fprintf(out, "  Disk Size: %d GB\n", cfg.Isolation.Defaults.VM.DiskSize)
	fmt.Fprintf(out, "  Base Image: %s\n", cfg.Isolation.Defaults.VM.BaseImage)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "GitHub:")
	fmt.Fprintf(out, "  Default Branch Prefix: %s\n", cfg.Isolation.Defaults.GitHub.DefaultBranchPrefix)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Output:")
	fmt.Fprintf(out, "  Sync Directory: %s\n", cfg.Isolation.Defaults.Output.SyncDir)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Proxy:")
	fmt.Fprintf(out, "  Mode: %s\n", cfg.Isolation.Defaults.Proxy.Mode)
	fmt.Fprintln(out)

	if vmName != "" {
		fmt.Fprintf(out, "(Showing config for VM: %s)\n", vmName)
	} else {
		fmt.Fprintln(out, "(Showing global config)")
	}

	return nil
}
