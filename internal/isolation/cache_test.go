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
	t.Run("when home dir is available should create homebrew cache directory", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		err := cm.SetupHomebrewCache()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act — first call
		if err := cm.SetupHomebrewCache(); err != nil {
			t.Fatalf("first SetupHomebrewCache failed: %v", err)
		}
		// Act — second call (idempotency check)
		err := cm.SetupHomebrewCache()

		// Assert
		if err != nil {
			t.Fatalf("second SetupHomebrewCache failed: %v", err)
		}
	})
}

func TestGetHomebrewCacheInfo(t *testing.T) {
	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		info, err := cm.GetHomebrewCacheInfo()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
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
		info, err := cm.GetHomebrewCacheInfo()

		// Assert
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
	t.Run("when homebrew cache is set up should include homebrew in output", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupHomebrewCache(); err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}
		var buf bytes.Buffer

		// Act
		err := cm.Status(&buf)

		// Assert
		if err != nil {
			t.Fatalf("Status failed: %v", err)
		}
		output := buf.String()
		if output == "" {
			t.Fatalf("expected non-empty status output")
		}
		if !strings.Contains(output, "Homebrew") {
			t.Fatalf("expected 'Homebrew' in status output")
		}
	})
}


func TestNewCacheManagerWithDirs(t *testing.T) {
	t.Run("when dirs provided should create homebrew cache under cache base dir", func(t *testing.T) {
		// Arrange
		homeDir := t.TempDir()
		cacheBaseDir := filepath.Join(homeDir, "cache")

		// Act
		cm := NewCacheManagerWithDirs(homeDir, cacheBaseDir)
		err := cm.SetupHomebrewCache()

		// Assert
		if err != nil {
			t.Fatalf("expected no error setting up homebrew cache, got: %v", err)
		}
		expectedDir := filepath.Join(cacheBaseDir, "homebrew")
		if _, statErr := os.Stat(expectedDir); os.IsNotExist(statErr) {
			t.Fatalf("expected homebrew cache dir at %q, but it does not exist", expectedDir)
		}
	})
}

func TestHomebrewCacheSetupEdgeCases(t *testing.T) {
	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		err := cm.SetupHomebrewCache()

		// Assert
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})

	t.Run("when cache base dir is read only should return permission error", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		readonlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.Mkdir(readonlyDir, 0444); err != nil {
			t.Fatalf("failed to create readonly dir: %v", err)
		}
		defer os.Chmod(readonlyDir, 0755)
		cm := NewCacheManagerWithDirs(tmpDir, readonlyDir)

		// Act
		err := cm.SetupHomebrewCache()

		// Assert
		if err == nil {
			t.Fatalf("expected error for permission denied, got nil")
		}
	})
}

func TestVMHomebrewCacheSetup(t *testing.T) {
	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupHomebrewCache(); err != nil {
			t.Fatalf("SetupHomebrewCache failed: %v", err)
		}

		// Act
		commands := cm.SetupVMHomebrewCache()

		// Assert
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}
		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		commands := cm.SetupVMHomebrewCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		// Act
		commands := cm.SetupVMHomebrewCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestSharedCacheMountAndHostPath(t *testing.T) {
	t.Run("when called should return correct mount specification", func(t *testing.T) {
		// Arrange
		cm := NewCacheManager()

		// Act
		mount := cm.GetSharedCacheMount()

		// Assert
		expected := "calf-cache:~/.calf-cache"
		if mount != expected {
			t.Fatalf("expected mount spec %s, got %s", expected, mount)
		}
	})

	t.Run("when home dir is available should return path with calf-cache prefix", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		hostPath := cm.GetHomebrewCacheHostPath()

		// Assert
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
	t.Run("when bytes is zero should return 0 B", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(0)

		// Assert
		if result != "0 B" {
			t.Errorf("FormatBytes(0) = %s, expected 0 B", result)
		}
	})

	t.Run("when bytes is under one KB should display raw bytes", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(512)

		// Assert
		if result != "512 B" {
			t.Errorf("FormatBytes(512) = %s, expected 512 B", result)
		}
	})

	t.Run("when bytes is exactly one KB should display 1.0 KB", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(1024)

		// Assert
		if result != "1.0 KB" {
			t.Errorf("FormatBytes(1024) = %s, expected 1.0 KB", result)
		}
	})

	t.Run("when bytes is fractional KB should display with one decimal", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(1536)

		// Assert
		if result != "1.5 KB" {
			t.Errorf("FormatBytes(1536) = %s, expected 1.5 KB", result)
		}
	})

	t.Run("when bytes is exactly one MB should display 1.0 MB", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(1048576)

		// Assert
		if result != "1.0 MB" {
			t.Errorf("FormatBytes(1048576) = %s, expected 1.0 MB", result)
		}
	})

	t.Run("when bytes is exactly one GB should display 1.0 GB", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(1073741824)

		// Assert
		if result != "1.0 GB" {
			t.Errorf("FormatBytes(1073741824) = %s, expected 1.0 GB", result)
		}
	})

	t.Run("when bytes is exactly one TB should display 1.0 TB", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(1099511627776)

		// Assert
		if result != "1.0 TB" {
			t.Errorf("FormatBytes(1099511627776) = %s, expected 1.0 TB", result)
		}
	})

	t.Run("when bytes is 5 GB should display 5.0 GB", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(5368709120)

		// Assert
		if result != "5.0 GB" {
			t.Errorf("FormatBytes(5368709120) = %s, expected 5.0 GB", result)
		}
	})

	t.Run("when bytes is fractional MB should display with one decimal", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		result := FormatBytes(2621440)

		// Assert
		if result != "2.5 MB" {
			t.Errorf("FormatBytes(2621440) = %s, expected 2.5 MB", result)
		}
	})
}

func TestNpmCacheSetup(t *testing.T) {
	t.Run("when home dir is available should create npm cache directory", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		err := cm.SetupNpmCache()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act — first call
		if err := cm.SetupNpmCache(); err != nil {
			t.Fatalf("first SetupNpmCache failed: %v", err)
		}
		// Act — second call (idempotency check)
		err := cm.SetupNpmCache()

		// Assert
		if err != nil {
			t.Fatalf("second SetupNpmCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		err := cm.SetupNpmCache()

		// Assert
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetNpmCacheInfo(t *testing.T) {
	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		info, err := cm.GetNpmCacheInfo()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupNpmCache(); err != nil {
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

		// Act
		info, err := cm.GetNpmCacheInfo()

		// Assert
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
	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupNpmCache(); err != nil {
			t.Fatalf("SetupNpmCache failed: %v", err)
		}

		// Act
		commands := cm.SetupVMNpmCache()

		// Assert
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}
		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		commands := cm.SetupVMNpmCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		// Act
		commands := cm.SetupVMNpmCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGoCacheSetup(t *testing.T) {
	t.Run("when home dir is available should create go cache directory with subdirs", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		err := cm.SetupGoCache()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act — first call
		if err := cm.SetupGoCache(); err != nil {
			t.Fatalf("first SetupGoCache failed: %v", err)
		}
		// Act — second call (idempotency check)
		err := cm.SetupGoCache()

		// Assert
		if err != nil {
			t.Fatalf("second SetupGoCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		err := cm.SetupGoCache()

		// Assert
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetGoCacheInfo(t *testing.T) {
	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		info, err := cm.GetGoCacheInfo()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGoCache(); err != nil {
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

		// Act
		info, err := cm.GetGoCacheInfo()

		// Assert
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
	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGoCache(); err != nil {
			t.Fatalf("SetupGoCache failed: %v", err)
		}

		// Act
		commands := cm.SetupVMGoCache()

		// Assert
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}
		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		commands := cm.SetupVMGoCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		// Act
		commands := cm.SetupVMGoCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGitCacheSetup(t *testing.T) {
	t.Run("when home dir is available should create git cache directory", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		err := cm.SetupGitCache()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act — first call
		if err := cm.SetupGitCache(); err != nil {
			t.Fatalf("first SetupGitCache failed: %v", err)
		}
		// Act — second call (idempotency check)
		err := cm.SetupGitCache()

		// Assert
		if err != nil {
			t.Fatalf("second SetupGitCache failed: %v", err)
		}
	})

	t.Run("when home dir is empty should return nil without error", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		err := cm.SetupGitCache()

		// Assert
		if err != nil {
			t.Fatalf("expected graceful degradation (nil error) when homeDir unavailable, got: %v", err)
		}
	})
}

func TestGetGitCacheInfo(t *testing.T) {
	t.Run("when cache does not exist should return zero size", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		info, err := cm.GetGitCacheInfo()

		// Assert
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
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGitCache(); err != nil {
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

		// Act
		info, err := cm.GetGitCacheInfo()

		// Assert
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
	t.Run("when host cache exists should return setup commands", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGitCache(); err != nil {
			t.Fatalf("SetupGitCache failed: %v", err)
		}

		// Act
		commands := cm.SetupVMGitCache()

		// Assert
		if commands == nil {
			t.Fatalf("expected non-nil commands")
		}
		if len(commands) == 0 {
			t.Fatalf("expected at least one command")
		}
	})

	t.Run("when home dir is unavailable should return nil", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		commands := cm.SetupVMGitCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when homeDir unavailable, got: %v", commands)
		}
	})

	t.Run("when host cache does not exist should return nil", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "nonexistent-cache"))

		// Act
		commands := cm.SetupVMGitCache()

		// Assert
		if commands != nil {
			t.Fatalf("expected nil commands when host cache doesn't exist, got: %v", commands)
		}
	})
}

func TestGetCachedGitRepos(t *testing.T) {
	t.Run("when no repos are cached should return empty list", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		repos, err := cm.GetCachedGitRepos()

		// Assert
		if err != nil {
			t.Fatalf("GetCachedGitRepos failed: %v", err)
		}
		if len(repos) != 0 {
			t.Fatalf("expected empty list, got %d repos", len(repos))
		}
	})

	t.Run("when repos are cached should return their names", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGitCache(); err != nil {
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

		// Act
		repos, err := cm.GetCachedGitRepos()

		// Assert
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
	t.Run("when repo already cached should return false", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGitCache(); err != nil {
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

		// Act
		created, err := cm.CacheGitRepo("https://example.com/repo.git", "test-repo")

		// Assert
		if err != nil {
			t.Fatalf("CacheGitRepo failed: %v", err)
		}
		if created {
			t.Fatalf("expected false when repo already exists, got true")
		}
	})

	t.Run("when home dir is unavailable should return error", func(t *testing.T) {
		// Arrange
		cm := NewCacheManagerWithDirs("", "")

		// Act
		_, err := cm.CacheGitRepo("https://example.com/repo.git", "test-repo")

		// Assert
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
	t.Run("when no repos are cached should return zero updates", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))

		// Act
		updated, err := cm.UpdateGitRepos()

		// Assert
		if err != nil {
			t.Fatalf("UpdateGitRepos failed: %v", err)
		}
		if updated != 0 {
			t.Fatalf("expected 0 updates when no repos cached, got %d", updated)
		}
	})

	t.Run("when directory is not a git repo should skip without error", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		cm := NewCacheManagerWithDirs(tmpDir, filepath.Join(tmpDir, "cache"))
		if err := cm.SetupGitCache(); err != nil {
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

		// Act
		updated, err := cm.UpdateGitRepos()

		// Assert
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
		// Arrange
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		// Act
		_ = cm.SetupNpmCache()

		// Assert
		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when home dir is empty should route go warning to injected writer", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		// Act
		_ = cm.SetupGoCache()

		// Assert
		if !strings.Contains(buf.String(), "Warning") {
			t.Fatalf("expected warning in injected writer, got: %q", buf.String())
		}
	})

	t.Run("when home dir is empty should route git warning to injected writer", func(t *testing.T) {
		// Arrange
		var buf bytes.Buffer
		cm := NewCacheManagerWithWriter("", "", &buf)

		// Act
		_ = cm.SetupGitCache()

		// Assert
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
		// Arrange — no setup needed

		// Act
		cm := NewCacheManager()

		// Assert
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
