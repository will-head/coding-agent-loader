package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/will-head/coding-agent-launcher/internal/isolation"
)

// newCacheCmd creates the cache command group with injectable stdin and homeDir.
// Pass an empty homeDir to use the default (current user's home directory).
func newCacheCmd(stdin io.Reader, homeDir string) *cobra.Command {
	cacheCmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage package download caches",
		Long:  `Manage package download caches for faster VM bootstraps.`,
	}

	cacheStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show cache status and sizes",
		Long:  `Display information about package download caches, including size, location, and availability.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := newCacheManager(homeDir)
			return cm.Status(cmd.OutOrStdout())
		},
	}

	var clearAll bool
	var force bool
	var dryRun bool
	var clearHomebrew bool
	var clearNpm bool
	var clearGo bool
	var clearGit bool

	cacheClearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear package download caches",
		Long: `Clear package download caches to free disk space.

By default, prompts for confirmation before clearing each cache type.
With --all, shows a final confirmation before clearing all caches.
With --all --force, skips all confirmations (for automation).
Use --homebrew, --npm, --go, or --git to clear a specific cache type.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cm := newCacheManager(homeDir)
			types := cacheTypeFlags{clearHomebrew, clearNpm, clearGo, clearGit}
			return runCacheClear(cmd, stdin, cm, clearAll, force, dryRun, types)
		},
	}

	cacheClearCmd.Flags().BoolVarP(&clearAll, "all", "a", false, "Clear all caches (requires confirmation unless --force is used)")
	cacheClearCmd.Flags().BoolVarP(&force, "force", "f", false, "Skip all confirmations (use with --all for automation)")
	cacheClearCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be cleared without actually clearing")
	cacheClearCmd.Flags().BoolVar(&clearHomebrew, "homebrew", false, "Clear only the Homebrew cache")
	cacheClearCmd.Flags().BoolVar(&clearNpm, "npm", false, "Clear only the npm cache")
	cacheClearCmd.Flags().BoolVar(&clearGo, "go", false, "Clear only the Go cache")
	cacheClearCmd.Flags().BoolVar(&clearGit, "git", false, "Clear only the Git cache")

	cacheCmd.AddCommand(cacheStatusCmd)
	cacheCmd.AddCommand(cacheClearCmd)

	return cacheCmd
}

// newCacheManager creates a CacheManager, using a custom homeDir when non-empty.
func newCacheManager(homeDir string) *isolation.CacheManager {
	if homeDir != "" {
		return isolation.NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, ".calf-cache"))
	}
	return isolation.NewCacheManager()
}

// cacheTypeFlags holds per-type clear flags.
type cacheTypeFlags struct {
	homebrew bool
	npm      bool
	goCache  bool
	git      bool
}

// anySet returns true if any per-type flag is set.
func (f cacheTypeFlags) anySet() bool {
	return f.homebrew || f.npm || f.goCache || f.git
}

// runCacheClear executes the cache clear operation using the provided command I/O, cache manager, and flags.
func runCacheClear(cmd *cobra.Command, stdin io.Reader, cm *isolation.CacheManager, clearAll, force, dryRun bool, types cacheTypeFlags) error {
	allCacheTypes := []struct {
		name     string
		clearKey string // lowercase key accepted by CacheManager.Clear
		flag     bool
		getInfo  func() (*isolation.CacheInfo, error)
	}{
		{"Homebrew", "homebrew", types.homebrew, cm.GetHomebrewCacheInfo},
		{"npm", "npm", types.npm, cm.GetNpmCacheInfo},
		{"Go", "go", types.goCache, cm.GetGoCacheInfo},
		{"Git", "git", types.git, cm.GetGitCacheInfo},
	}

	// When per-type flags are used, limit to only those types (treat as --all --force for those).
	if types.anySet() {
		for _, ct := range allCacheTypes {
			if !ct.flag {
				continue
			}
			info, err := ct.getInfo()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error getting %s cache info: %v\n", ct.name, err)
				continue
			}
			if !info.Available {
				fmt.Fprintf(cmd.OutOrStdout(), "No %s cache found\n", ct.name)
				continue
			}
			sizeStr := isolation.FormatBytes(info.Size)
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would clear %s cache (%s)\n", ct.name, sizeStr)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Clearing %s cache (%s)...\n", ct.name, sizeStr)
				if _, err := cm.Clear(ct.clearKey, false); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error clearing %s cache: %v\n", ct.name, err)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Cleared %s cache\n", ct.name)
				}
			}
		}
		return nil
	}

	totalFreed := int64(0)
	clearedCount := 0
	totalCount := 0

	// Create the reader once so the same buffer is used for both the global
	// confirmation prompt (clearAll) and the per-cache prompts below.
	reader := bufio.NewReader(stdin)

	if clearAll && !dryRun && !force {
		fmt.Fprintln(cmd.OutOrStdout(), "Warning: This will clear ALL caches!")
		fmt.Fprintln(cmd.OutOrStdout(), "This will slow down your next VM bootstrap.")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprint(cmd.OutOrStdout(), "Are you sure you want to continue? [y/N]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(cmd.OutOrStdout(), "\nAborted (EOF)")
				return nil
			}
			return fmt.Errorf("failed to read input: %w", err)
		}

		if strings.TrimSpace(strings.ToLower(input)) != "y" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted")
			return nil
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	if dryRun {
		fmt.Fprintln(cmd.OutOrStdout(), "Dry run: Showing what would be cleared")
		fmt.Fprintln(cmd.OutOrStdout())
	}

	for _, ct := range allCacheTypes {
		info, err := ct.getInfo()
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error getting %s cache info: %v\n", ct.name, err)
			continue
		}

		if !info.Available {
			continue
		}

		totalCount++

		sizeStr := isolation.FormatBytes(info.Size)
		var shouldClear bool

		if clearAll {
			shouldClear = true
			if dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would clear %s cache (%s)\n", ct.name, sizeStr)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Clearing %s cache (%s)...\n", ct.name, sizeStr)
			}
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "Clear %s cache (%s)? [y/N]: ", ct.name, sizeStr)
			input, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					shouldClear = false
					fmt.Fprintf(cmd.OutOrStdout(), "Skipping %s cache (EOF)\n", ct.name)
				} else {
					return fmt.Errorf("failed to read input: %w", err)
				}
			} else {
				shouldClear = strings.TrimSpace(strings.ToLower(input)) == "y"
			}
		}

		if shouldClear {
			cleared, err := cm.Clear(ct.clearKey, dryRun)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error clearing %s cache: %v\n", ct.name, err)
				continue
			}

			if cleared {
				totalFreed += info.Size
				clearedCount++
				if !clearAll && !dryRun {
					fmt.Fprintf(cmd.OutOrStdout(), "Cleared %s cache\n", ct.name)
				}
			}
		} else {
			if !clearAll {
				fmt.Fprintf(cmd.OutOrStdout(), "Skipping %s cache\n", ct.name)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	if totalCount == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No caches found to clear")
		return nil
	}

	if clearedCount == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No caches cleared")
		return nil
	}

	action := "Cleared"
	if dryRun {
		action = "Would clear"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "%s %s (%d/%d caches)\n", action, isolation.FormatBytes(totalFreed), clearedCount, totalCount)

	if !dryRun && clearedCount > 0 {
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Warning: Next VM bootstrap will be slower")
		fmt.Fprintln(cmd.OutOrStdout(), "Use 'calf cache status' to verify caches are empty")
	}

	return nil
}
