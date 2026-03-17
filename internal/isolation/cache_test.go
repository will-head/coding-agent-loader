// Package isolation provides VM isolation and management for CALF.
package isolation

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomebrewCacheSetup(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when home dir is available should create homebrew cache directory", func(t *testing.T) {
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}

		cacheInfo, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}
		if cacheInfo.Path == "" {
			t.Fatalf("expected non-empty cache path")
		}
		dirInfo, err := os.Stat(cacheInfo.Path)
		if err != nil {
			t.Fatalf("host cache directory not created: %v", err)
		}
		if !dirInfo.IsDir() {
			t.Fatalf("host cache is not a directory")
		}

		// Verify subdirectories exist
		downloadsDir := filepath.Join(cacheInfo.Path, "downloads")
		if _, err := os.Stat(downloadsDir); err != nil {
			t.Fatalf("downloads subdirectory not created: %v", err)
		}

		caskDir := filepath.Join(cacheInfo.Path, "Cask")
		if _, err := os.Stat(caskDir); err != nil {
			t.Fatalf("Cask subdirectory not created: %v", err)
		}
	})

	t.Run("when called twice should succeed both times", func(t *testing.T) {
		// First setup
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("first SetupHomebrewCache failed: %v", err)
		}

		// Second setup should not fail
		err = cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("second SetupHomebrewCache failed: %v", err)
		}
	})
}

func TestGetHomebrewCacheInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		info, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}

		if info.Size != 0 {
			t.Fatalf("expected size 0, got %d", info.Size)
		}

		if info.Path == "" {
			t.Fatalf("expected non-empty path")
		}
	})

	t.Run("when cache contains files should return non-zero size", func(t *testing.T) {
		// Setup cache
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}

		// Create a test file
		hbInfo, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(hbInfo.Path, "test-file.bin")
		if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		info, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}

		if info.Size == 0 {
			t.Fatalf("expected non-zero size")
		}

		if !info.Available {
			t.Fatalf("expected cache to be available")
		}
	})
}

func TestCacheStatus(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when homebrew cache is set up should include homebrew in output", func(t *testing.T) {
		// Setup cache
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}

		var buf bytes.Buffer

		err = cm.Status(&buf)
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Fatalf("expected non-empty status output")
		}

		// Verify cache info is present
		if !strings.Contains(output, "Homebrew") {
			t.Fatalf("expected 'Homebrew' in status output")
		}
	})
}


func TestNewCacheManagerWithDirs(t *testing.T) {
	t.Run("when dirs provided should initialise with given home and cache base dirs", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cacheBaseDir := filepath.Join(homeDir, "cache")

		// Act
		cm := NewCacheManagerWithDirs(homeDir, cacheBaseDir)

		// Assert
		if cm == nil {
			t.Fatalf("expected non-nil CacheManager")
		}
		if cm.homeDir != homeDir {
			t.Fatalf("expected homeDir %q, got %q", homeDir, cm.homeDir)
		}
		if cm.cacheBaseDir != cacheBaseDir {
			t.Fatalf("expected cacheBaseDir %q, got %q", cacheBaseDir, cm.cacheBaseDir)
		}
	})
}

func TestHomebrewCacheSetupEdgeCases(t *testing.T) {
	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		cm := NewCacheManagerWithDirs("", "")

		// Should not return error when home directory unavailable (graceful degradation)
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})

	t.Run("when cache base dir is read only should return permission error", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a directory with read-only permissions
		readonlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.Mkdir(readonlyDir, 0444); err != nil {
			t.Fatalf("failed to create readonly dir: %v", err)
		}
		defer os.Chmod(readonlyDir, 0755)

		cm := NewCacheManagerWithDirs(tmpDir, readonlyDir)

		// Should return error for permission issues (not graceful degradation case)
		err = cm.SetupHomebrewCache()
		if err == nil {
			t.Fatalf("expected error for permission denied, got nil")
		}
	})
}

func TestVMHomebrewCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		// Setup host cache
		err := cm.SetupHomebrewCache()
		if err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}

		commands := cm.SetupVMHomebrewCache()
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}

		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}

		// Verify commands contain expected operations (mount verification, not symlinks)
		commandsStr := strings.Join(commands, " ")
		if !strings.Contains(commandsStr, "mount | grep -q \" on $HOME/.calf-cache \"") {
			t.Fatalf("expected mount verification in VM setup")
		}
		if !strings.Contains(commandsStr, "test -d") {
			t.Fatalf("expected cache directory verification in VM setup")
		}
		if !strings.Contains(commandsStr, "HOMEBREW_CACHE") {
			t.Fatalf("expected HOMEBREW_CACHE environment variable setup")
		}
		if !strings.Contains(commandsStr, "touch ~/.zshrc") {
			t.Fatalf("expected touch command to ensure .zshrc exists before grep")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		commands := cmNoHome.SetupVMHomebrewCache()
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		cmNoCache := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		commands := cmNoCache.SetupVMHomebrewCache()
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestSharedCacheMountAndHostPath(t *testing.T) {
	t.Run("when called should return correct mount specification", func(t *testing.T) {
		cm := NewCacheManager()
		mount := cm.GetSharedCacheMount()

		expected := "calf-cache:~/.calf-cache"
		if mount != expected {
			t.Fatalf("expected mount spec %s, got %s", expected, mount)
		}
	})

	t.Run("when home dir is available should return path with calf-cache prefix", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		hostPath := cm.GetHomebrewCacheHostPath()
		if hostPath == "" {
			t.Fatalf("expected non-empty host path")
		}

		if !strings.Contains(hostPath, "calf-cache:") {
			t.Fatalf("expected 'calf-cache:' prefix in host path")
		}

		if !strings.Contains(hostPath, "homebrew") {
			t.Fatalf("expected 'homebrew' in host path")
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"when bytes is zero should return 0 B", 0, "0 B"},
		{"when bytes is under one KB should display raw bytes", 512, "512 B"},
		{"when bytes is exactly one KB should display 1.0 KB", 1024, "1.0 KB"},
		{"when bytes is fractional KB should display with one decimal", 1536, "1.5 KB"},
		{"when bytes is exactly one MB should display 1.0 MB", 1048576, "1.0 MB"},
		{"when bytes is exactly one GB should display 1.0 GB", 1073741824, "1.0 GB"},
		{"when bytes is exactly one TB should display 1.0 TB", 1099511627776, "1.0 TB"},
		{"when bytes is 5 GB should display 5.0 GB", 5368709120, "5.0 GB"},
		{"when bytes is fractional MB should display with one decimal", 2621440, "2.5 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestNpmCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when home dir is available should create npm cache directory", func(t *testing.T) {
		err := cm.SetupNpmCache()
		if err != nil {
			t.Fatalf("SetupNpmCache failed: %v", err)
		}

		cacheInfo, err := cm.GetNpmCacheInfo()
		if err != nil {
			t.Fatalf("GetNpmCacheInfo failed: %v", err)
		}
		dirInfo, err := os.Stat(cacheInfo.Path)
		if err != nil {
			t.Fatalf("host cache directory not created: %v", err)
		}
		if !dirInfo.IsDir() {
			t.Fatalf("host cache is not a directory")
		}
	})

	t.Run("when called twice should succeed both times", func(t *testing.T) {
		err := cm.SetupNpmCache()
		if err != nil {
			t.Fatalf("first SetupNpmCache failed: %v", err)
		}

		err = cm.SetupNpmCache()
		if err != nil {
			t.Fatalf("second SetupNpmCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		err := cmNoHome.SetupNpmCache()
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetNpmCacheInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		info, err := cm.GetNpmCacheInfo()
		if err != nil {
			t.Fatalf("GetNpmCacheInfo failed: %v", err)
		}

		if info.Size != 0 {
			t.Fatalf("expected size 0, got %d", info.Size)
		}

		if info.Path == "" {
			t.Fatalf("expected non-empty path")
		}

		if info.Available {
			t.Fatalf("expected cache to be unavailable")
		}
	})

	t.Run("when cache contains files should return non-zero size", func(t *testing.T) {
		err := cm.SetupNpmCache()
		if err != nil {
			t.Fatalf("SetupNpmCache failed: %v", err)
		}

		npmInfo, err := cm.GetNpmCacheInfo()
		if err != nil {
			t.Fatalf("GetNpmCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(npmInfo.Path, "test-package.tar.gz")
		if err := os.WriteFile(testFile, []byte("test npm cache data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		info, err := cm.GetNpmCacheInfo()
		if err != nil {
			t.Fatalf("GetNpmCacheInfo failed: %v", err)
		}

		if info.Size == 0 {
			t.Fatalf("expected non-zero size")
		}

		if !info.Available {
			t.Fatalf("expected cache to be available")
		}
	})
}

func TestVMNpmCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		err := cm.SetupNpmCache()
		if err != nil {
			t.Fatalf("SetupNpmCache failed: %v", err)
		}

		commands := cm.SetupVMNpmCache()
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}

		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}

		commandsStr := strings.Join(commands, " ")
		if !strings.Contains(commandsStr, "mount | grep -q \" on $HOME/.calf-cache \"") {
			t.Fatalf("expected mount verification in VM setup")
		}
		if !strings.Contains(commandsStr, "test -d") {
			t.Fatalf("expected cache directory verification in VM setup")
		}
		if !strings.Contains(commandsStr, "npm config set cache") {
			t.Fatalf("expected npm cache configuration")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		commands := cmNoHome.SetupVMNpmCache()
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		cmNoCache := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		commands := cmNoCache.SetupVMNpmCache()
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGoCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when home dir is available should create go cache directory with subdirs", func(t *testing.T) {
		err := cm.SetupGoCache()
		if err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}

		cacheInfo, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}
		if cacheInfo.Path == "" {
			t.Fatalf("expected non-empty cache path")
		}
		if !cacheInfo.Available {
			t.Fatalf("expected cache to be available after setup")
		}
	})

	t.Run("when called twice should succeed both times", func(t *testing.T) {
		err := cm.SetupGoCache()
		if err != nil {
			t.Fatalf("first SetupGoCache failed: %v", err)
		}

		err = cm.SetupGoCache()
		if err != nil {
			t.Fatalf("second SetupGoCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		err := cmNoHome.SetupGoCache()
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetGoCacheInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		info, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}

		if info.Size != 0 {
			t.Fatalf("expected size 0, got %d", info.Size)
		}

		if info.Path == "" {
			t.Fatalf("expected non-empty path")
		}

		if info.Available {
			t.Fatalf("expected cache to be unavailable")
		}
	})

	t.Run("when cache contains files should return non-zero size", func(t *testing.T) {
		err := cm.SetupGoCache()
		if err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}

		goInfo, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(goInfo.Path, "pkg", "mod", "test-module@v1.0.0")
		if err := os.WriteFile(testFile, []byte("test go module data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		info, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}

		if info.Size == 0 {
			t.Fatalf("expected non-zero size")
		}

		if !info.Available {
			t.Fatalf("expected cache to be available")
		}
	})
}

func TestVMGoCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		err := cm.SetupGoCache()
		if err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}

		commands := cm.SetupVMGoCache()
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}

		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}

		commandsStr := strings.Join(commands, " ")
		if !strings.Contains(commandsStr, "mount | grep -q \" on $HOME/.calf-cache \"") {
			t.Fatalf("expected mount verification in VM setup")
		}
		if !strings.Contains(commandsStr, "test -d") {
			t.Fatalf("expected cache directory verification in VM setup")
		}
		if !strings.Contains(commandsStr, "GOMODCACHE") {
			t.Fatalf("expected GOMODCACHE environment variable setup")
		}
		if !strings.Contains(commandsStr, "touch ~/.zshrc") {
			t.Fatalf("expected touch command to ensure .zshrc exists before grep")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		commands := cmNoHome.SetupVMGoCache()
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		cmNoCache := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		commands := cmNoCache.SetupVMGoCache()
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGitCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when home dir is available should create git cache directory", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		cacheInfo, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		dirInfo, err := os.Stat(cacheInfo.Path)
		if err != nil {
			t.Fatalf("host cache directory not created: %v", err)
		}
		if !dirInfo.IsDir() {
			t.Fatalf("host cache is not a directory")
		}
	})

	t.Run("when called twice should succeed both times", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("first SetupGitCache failed: %v", err)
		}

		err = cm.SetupGitCache()
		if err != nil {
			t.Fatalf("second SetupGitCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		err := cmNoHome.SetupGitCache()
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetGitCacheInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		info, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}

		if info.Size != 0 {
			t.Fatalf("expected size 0, got %d", info.Size)
		}

		if info.Path == "" {
			t.Fatalf("expected non-empty path")
		}

		if info.Available {
			t.Fatalf("expected cache to be unavailable")
		}
	})

	t.Run("when cache contains files should return non-zero size", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		gitInfo, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		repoCacheDir := filepath.Join(gitInfo.Path, "test-repo")
		if err := os.MkdirAll(repoCacheDir, 0755); err != nil {
			t.Fatalf("failed to create test repo cache: %v", err)
		}
		testFile := filepath.Join(repoCacheDir, "test-file.bin")
		if err := os.WriteFile(testFile, []byte("test git cache data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		info, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}

		if info.Size == 0 {
			t.Fatalf("expected non-zero size")
		}

		if !info.Available {
			t.Fatalf("expected cache to be available")
		}
	})
}

func TestVMGitCacheSetup(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		commands := cm.SetupVMGitCache()
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}

		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}

		commandsStr := strings.Join(commands, " ")
		if !strings.Contains(commandsStr, "mount | grep -q \" on $HOME/.calf-cache \"") {
			t.Fatalf("expected mount verification in VM setup")
		}
		if !strings.Contains(commandsStr, "test -d") {
			t.Fatalf("expected cache directory verification in VM setup")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		commands := cmNoHome.SetupVMGitCache()
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		cmNoCache := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		commands := cmNoCache.SetupVMGitCache()
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGetCachedGitRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when no repos are cached should return empty list", func(t *testing.T) {
		repos, err := cm.GetCachedGitRepos()
		if err != nil {
			t.Fatalf("GetCachedGitRepos failed: %v", err)
		}

		if len(repos) != 0 {
			t.Fatalf("expected empty list, got %d repos", len(repos))
		}
	})

	t.Run("when repos are cached should return their names", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		gitInfo, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		repo1Dir := filepath.Join(gitInfo.Path, "repo1")
		repo2Dir := filepath.Join(gitInfo.Path, "repo2")
		if err := os.MkdirAll(repo1Dir, 0755); err != nil {
			t.Fatalf("failed to create repo1: %v", err)
		}
		if err := os.MkdirAll(repo2Dir, 0755); err != nil {
			t.Fatalf("failed to create repo2: %v", err)
		}

		repos, err := cm.GetCachedGitRepos()
		if err != nil {
			t.Fatalf("GetCachedGitRepos failed: %v", err)
		}

		if len(repos) != 2 {
			t.Fatalf("expected 2 repos, got %d", len(repos))
		}

		foundRepo1 := false
		foundRepo2 := false
		for _, repo := range repos {
			if repo == "repo1" {
				foundRepo1 = true
			}
			if repo == "repo2" {
				foundRepo2 = true
			}
		}

		if !foundRepo1 || !foundRepo2 {
			t.Fatalf("expected to find repo1 and repo2, got %v", repos)
		}
	})
}

func TestCacheGitRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when repo already cached should return false", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		gitInfo, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		repoDir := filepath.Join(gitInfo.Path, "test-repo")
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("failed to create test repo: %v", err)
		}

		created, err := cm.CacheGitRepo("https://example.com/repo.git", "test-repo")
		if err != nil {
			t.Fatalf("CacheGitRepo failed: %v", err)
		}

		if created {
			t.Fatalf("expected false when repo already exists, got true")
		}
	})

	t.Run("when home dir is unavailable should return error", func(t *testing.T) {
		cmNoHome := NewCacheManagerWithDirs("", "")

		_, err := cmNoHome.CacheGitRepo("https://example.com/repo.git", "test-repo")
		if err == nil {
			t.Fatalf("expected error when homeDir unavailable")
		}
	})
}

// makeBadGitRepo creates a git repo at dir with a remote pointing to a non-existent path,
// causing git fetch to fail. dir must already exist or will be created.
func makeBadGitRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := exec.Command("git", "init", dir).Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	if err := exec.Command("git", "-C", dir, "remote", "add", "origin", "/nonexistent/path").Run(); err != nil {
		t.Fatalf("git remote add failed: %v", err)
	}
}

func TestUpdateGitRepos(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "calf-cache-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

	t.Run("when no repos are cached should return zero updates", func(t *testing.T) {
		updated, err := cm.UpdateGitRepos()
		if err != nil {
			t.Fatalf("UpdateGitRepos failed: %v", err)
		}

		if updated != 0 {
			t.Fatalf("expected 0 updates when no repos cached, got %d", updated)
		}
	})

	t.Run("when directory is not a git repo should skip without error", func(t *testing.T) {
		err := cm.SetupGitCache()
		if err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		gitInfo, err := cm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		repoDir := filepath.Join(gitInfo.Path, "not-a-repo")
		if err := os.MkdirAll(repoDir, 0755); err != nil {
			t.Fatalf("failed to create non-repo directory: %v", err)
		}

		updated, err := cm.UpdateGitRepos()
		if err != nil {
			t.Fatalf("UpdateGitRepos failed: %v", err)
		}

		if updated != 0 {
			t.Fatalf("expected 0 updates for non-git directory, got %d", updated)
		}
	})

	t.Run("when all repos update successfully should return updated count and nil error", func(t *testing.T) {
		// Arrange — clone a local bare repo into the git cache
		homeDir := t.TempDir()
		localCm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := localCm.SetupGitCache(); err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}
		bareRepo := filepath.Join(homeDir, "source.git")
		if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
			t.Fatalf("git init --bare failed: %v", err)
		}
		localGitInfo, err := localCm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		cacheRepoDir := filepath.Join(localGitInfo.Path, "source-repo")
		if err := exec.Command("git", "clone", bareRepo, cacheRepoDir).Run(); err != nil {
			t.Fatalf("git clone failed: %v", err)
		}

		// Act
		updated, err := localCm.UpdateGitRepos()

		// Assert
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if updated != 1 {
			t.Fatalf("expected 1 updated repo, got %d", updated)
		}
	})

	t.Run("when a repo fetch fails should return error", func(t *testing.T) {
		// Arrange — git repo with a non-existent remote
		homeDir := t.TempDir()
		localCm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := localCm.SetupGitCache(); err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}
		localGitInfo, err := localCm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		makeBadGitRepo(t, filepath.Join(localGitInfo.Path, "bad-repo"))

		// Act
		_, err = localCm.UpdateGitRepos()

		// Assert
		if err == nil {
			t.Fatal("expected error when repo fetch fails, got nil")
		}
	})

	t.Run("when one repo fails should attempt remaining repos and return partial count with error", func(t *testing.T) {
		// Arrange — bad repo (alphabetically first) + good repo (alphabetically last)
		homeDir := t.TempDir()
		localCm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := localCm.SetupGitCache(); err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}
		localGitInfo, err := localCm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		makeBadGitRepo(t, filepath.Join(localGitInfo.Path, "a-bad-repo"))
		bareRepo := filepath.Join(homeDir, "source.git")
		if err := exec.Command("git", "init", "--bare", bareRepo).Run(); err != nil {
			t.Fatalf("git init --bare failed: %v", err)
		}
		goodRepoDir := filepath.Join(localGitInfo.Path, "z-good-repo")
		if err := exec.Command("git", "clone", bareRepo, goodRepoDir).Run(); err != nil {
			t.Fatalf("git clone failed: %v", err)
		}

		// Act
		updated, err := localCm.UpdateGitRepos()

		// Assert — good repo was still reached and counted despite earlier failure
		if updated != 1 {
			t.Fatalf("expected 1 updated repo (bad one skipped, good one reached), got %d", updated)
		}
		if err == nil {
			t.Fatal("expected error reflecting failed repo, got nil")
		}
	})
}

func TestCacheManagerWriterInjection(t *testing.T) {
	t.Run("when home dir is empty should write warning to injected writer not stderr", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		// Act
		err := cm.SetupHomebrewCache()

		// Assert
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when home dir is empty should route npm warning to injected writer", func(t *testing.T) {
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		_ = cm.SetupNpmCache()

		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when home dir is empty should route go warning to injected writer", func(t *testing.T) {
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		_ = cm.SetupGoCache()

		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when home dir is empty should route git warning to injected writer", func(t *testing.T) {
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		_ = cm.SetupGitCache()

		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when repo fetch fails should route warning to injected writer", func(t *testing.T) {
		// Arrange — bad repo so UpdateGitRepos produces a per-repo warning
		homeDir := t.TempDir()
		var buf bytes.Buffer
		localCm := NewCacheManagerWithWriter(homeDir, filepath.Join(homeDir, "cache"), &buf)
		if err := localCm.SetupGitCache(); err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}
		localGitInfo, err := localCm.GetGitCacheInfo()
		if err != nil {
			t.Fatalf("GetGitCacheInfo failed: %v", err)
		}
		makeBadGitRepo(t, filepath.Join(localGitInfo.Path, "bad-repo"))

		// Act
		_, _ = localCm.UpdateGitRepos()

		// Assert
		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when no writer injected should default to os.Stderr without panic", func(t *testing.T) {
		// NewCacheManager() must not panic — it defaults writer to os.Stderr
		cm := NewCacheManager()
		if cm == nil {
			t.Fatal("expected non-nil CacheManager")
		}
	})
}

func TestClearCache(t *testing.T) {
	t.Run("when cache has files should delete files and recreate directory", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := cm.SetupHomebrewCache(); err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}
		hbInfo, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(hbInfo.Path, "test-file.bin")
		if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Act
		cleared, err := cm.Clear("homebrew", false)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !cleared {
			t.Fatalf("expected cleared=true, got false")
		}
		if _, err := os.Stat(testFile); !os.IsNotExist(err) {
			t.Fatalf("expected test file to be deleted, but it still exists")
		}
		dirInfo, err := os.Stat(hbInfo.Path)
		if err != nil {
			t.Fatalf("expected cache directory to be recreated: %v", err)
		}
		if !dirInfo.IsDir() {
			t.Fatalf("expected cache directory to be a directory")
		}
	})

	t.Run("when dry run is true should not delete files", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := cm.SetupHomebrewCache(); err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}
		hbInfo, err := cm.GetHomebrewCacheInfo()
		if err != nil {
			t.Fatalf("GetHomebrewCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(hbInfo.Path, "test-file.bin")
		if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Act
		cleared, err := cm.Clear("homebrew", true)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !cleared {
			t.Fatalf("expected cleared=true in dry run mode")
		}
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Fatalf("expected test file to exist in dry run mode, but it was deleted")
		}
	})

	t.Run("when cache does not exist should return false", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))

		// Act
		cleared, err := cm.Clear("homebrew", false)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if cleared {
			t.Fatalf("expected cleared=false when cache doesn't exist")
		}
	})

	t.Run("when cache type is valid should clear that cache type", func(t *testing.T) {
		testCases := []string{"homebrew", "npm", "go", "git"}

		for _, cacheType := range testCases {
			t.Run(cacheType, func(t *testing.T) {
				// Arrange
				homeDir := t.TempDir()
				cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
				setupFuncs := map[string]func() error{
					"homebrew": cm.SetupHomebrewCache,
					"npm":      cm.SetupNpmCache,
					"go":       cm.SetupGoCache,
					"git":      cm.SetupGitCache,
				}
				setup, ok := setupFuncs[cacheType]
				if !ok {
					t.Fatalf("no setup func registered for cache type %s", cacheType)
				}
				if err := setup(); err != nil {
					t.Fatalf("Setup for %s failed: %v", cacheType, err)
				}

				// Act
				cleared, err := cm.Clear(cacheType, false)

				// Assert
				if err != nil {
					t.Fatalf("Clear failed for %s: %v", cacheType, err)
				}
				if !cleared {
					t.Fatalf("expected cleared=true for %s", cacheType)
				}
			})
		}
	})

	t.Run("when go cache is cleared should recreate pkg mod subdirectory", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := cm.SetupGoCache(); err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}
		goInfo, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}
		testFile := filepath.Join(goInfo.Path, "pkg", "mod", "test-file.bin")
		if err := os.WriteFile(testFile, []byte("test data"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Act
		cleared, err := cm.Clear("go", false)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !cleared {
			t.Fatalf("expected cleared=true, got false")
		}
		pkgModDir := filepath.Join(goInfo.Path, "pkg", "mod")
		if _, err := os.Stat(pkgModDir); err != nil {
			t.Fatalf("expected pkg/mod subdirectory to be recreated: %v", err)
		}
	})

	t.Run("when go cache has read only files should clear successfully", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cm := NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, "cache"))
		if err := cm.SetupGoCache(); err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}
		goInfo, err := cm.GetGoCacheInfo()
		if err != nil {
			t.Fatalf("GetGoCacheInfo failed: %v", err)
		}
		modDir := filepath.Join(goInfo.Path, "pkg", "mod")
		testModuleDir := filepath.Join(modDir, "gopkg.in", "yaml.v3@v3.0.1")
		if err := os.MkdirAll(testModuleDir, 0755); err != nil {
			t.Fatalf("failed to create test module directory: %v", err)
		}
		readOnlyFile := filepath.Join(testModuleDir, "decode_test.go")
		if err := os.WriteFile(readOnlyFile, []byte("package yaml"), 0444); err != nil {
			t.Fatalf("failed to create read-only test file: %v", err)
		}
		if err := os.Chmod(testModuleDir, 0555); err != nil {
			t.Fatalf("failed to make test module directory read-only: %v", err)
		}
		parentDir := filepath.Join(modDir, "gopkg.in")
		if err := os.Chmod(parentDir, 0555); err != nil {
			t.Fatalf("failed to make parent directory read-only: %v", err)
		}
		defer func() {
			os.Chmod(parentDir, 0755)
			os.Chmod(testModuleDir, 0755)
		}()

		// Act
		cleared, err := cm.Clear("go", false)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed with read-only files: %v", err)
		}
		if !cleared {
			t.Fatalf("expected cleared=true with read-only files")
		}
		if _, err := os.Stat(modDir); err != nil {
			t.Fatalf("expected pkg/mod subdirectory to be recreated: %v", err)
		}
		if _, err := os.Stat(readOnlyFile); !os.IsNotExist(err) {
			t.Fatalf("expected read-only test file to be deleted")
		}
	})

	t.Run("when cache is a symlink should preserve symlink and clear target contents", func(t *testing.T) {
		// Arrange — simulate the VM scenario where ~/.calf-cache/{type} is a symlink
		// to /Volumes/My Shared Files/calf-cache/{type}
		sharedVolumeRoot := t.TempDir()
		sharedCacheDir := filepath.Join(sharedVolumeRoot, "shared-volume", "npm")
		if err := os.MkdirAll(sharedCacheDir, 0755); err != nil {
			t.Fatalf("failed to create shared cache dir: %v", err)
		}
		testFile1 := filepath.Join(sharedCacheDir, "package1.tgz")
		testFile2 := filepath.Join(sharedCacheDir, "package2.tgz")
		if err := os.WriteFile(testFile1, []byte("package data 1"), 0644); err != nil {
			t.Fatalf("failed to create test file 1: %v", err)
		}
		if err := os.WriteFile(testFile2, []byte("package data 2"), 0644); err != nil {
			t.Fatalf("failed to create test file 2: %v", err)
		}

		vmHomeDir := t.TempDir()
		vmCm := NewCacheManagerWithDirs(vmHomeDir, filepath.Join(vmHomeDir, ".calf-cache"))
		npmInfo, err := vmCm.GetNpmCacheInfo()
		if err != nil {
			t.Fatalf("GetNpmCacheInfo failed: %v", err)
		}
		symlinkPath := npmInfo.Path
		if err := os.MkdirAll(filepath.Dir(symlinkPath), 0755); err != nil {
			t.Fatalf("failed to create .calf-cache dir: %v", err)
		}
		if err := os.Symlink(sharedCacheDir, symlinkPath); err != nil {
			t.Fatalf("failed to create symlink: %v", err)
		}
		expectedTarget, err := filepath.EvalSymlinks(sharedCacheDir)
		if err != nil {
			t.Fatalf("failed to resolve sharedCacheDir: %v", err)
		}

		// Act
		cleared, err := vmCm.Clear("npm", false)

		// Assert
		if err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
		if !cleared {
			t.Fatalf("expected cleared=true")
		}
		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("symlink was removed: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("expected symlink to be preserved, but it's now a regular directory")
		}
		newTarget, err := filepath.EvalSymlinks(symlinkPath)
		if err != nil {
			t.Fatalf("failed to resolve symlink after clear: %v", err)
		}
		if newTarget != expectedTarget {
			t.Fatalf("symlink target changed: got %s, want %s", newTarget, expectedTarget)
		}
		if _, err := os.Stat(testFile1); !os.IsNotExist(err) {
			t.Fatalf("expected test file 1 to be deleted")
		}
		if _, err := os.Stat(testFile2); !os.IsNotExist(err) {
			t.Fatalf("expected test file 2 to be deleted")
		}
		entries, err := os.ReadDir(sharedCacheDir)
		if err != nil {
			t.Fatalf("failed to read shared cache dir: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("expected empty directory, found %d entries", len(entries))
		}
	})
}
