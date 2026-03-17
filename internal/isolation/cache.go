// Package isolation provides VM isolation and management for CALF.
package isolation

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CacheManager manages package download caches for CALF VMs.
type CacheManager struct {
	homeDir      string
	cacheBaseDir string
}

// CacheInfo contains information about a cache.
type CacheInfo struct {
	// Path is the filesystem path to the cache directory.
	Path string
	// Size is the total size of the cache in bytes.
	Size int64
	// Available indicates whether the cache is configured and ready to use.
	Available bool
	// LastAccess is the last modification time of the cache directory.
	LastAccess time.Time
}

const (
	// homebrewCacheDir is the directory name for Homebrew cache under .calf-cache.
	homebrewCacheDir = "homebrew"
	// homebrewDownloadsDir is the subdirectory for Homebrew package downloads.
	homebrewDownloadsDir = "downloads"
	// homebrewCaskDir is the subdirectory for Homebrew Cask downloads.
	homebrewCaskDir = "Cask"
	// npmCacheDir is the directory name for npm cache under .calf-cache.
	npmCacheDir = "npm"
	// goCacheDir is the directory name for Go cache under .calf-cache.
	goCacheDir = "go"
	// gitCacheDir is the directory name for git cache under .calf-cache.
	gitCacheDir = "git"
	// sharedCacheMount is the Tart directory mount specification for cache sharing.
	sharedCacheMount = "calf-cache:~/.calf-cache"
)

// getDiskUsage returns the disk usage in bytes for a path using du -sk.
// Returns 0 if path doesn't exist or on error.
func getDiskUsage(path string) int64 {
	cmd := exec.Command("du", "-sk", path)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return 0
	}

	size, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}

	// du -sk returns size in kilobytes, convert to bytes
	return size * 1024
}

// NewCacheManager creates a new CacheManager with default paths.
func NewCacheManager() *CacheManager {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return NewCacheManagerWithDirs(homeDir, filepath.Join(homeDir, ".calf-cache"))
}

// NewCacheManagerWithDirs creates a CacheManager rooted at the given home and
// cache base directories. Intended for use in tests.
func NewCacheManagerWithDirs(homeDir, cacheBaseDir string) *CacheManager {
	return &CacheManager{
		homeDir:      homeDir,
		cacheBaseDir: cacheBaseDir,
	}
}

// getHomebrewCachePath returns the host path for Homebrew cache.
func (c *CacheManager) getHomebrewCachePath() string {
	return filepath.Join(c.cacheBaseDir, homebrewCacheDir)
}

// GetSharedCacheMount returns the Tart directory mount specification for cache sharing.
func (c *CacheManager) GetSharedCacheMount() string {
	return sharedCacheMount
}

// GetHomebrewCacheHostPath returns the host path for Homebrew cache mounting.
func (c *CacheManager) GetHomebrewCacheHostPath() string {
	return fmt.Sprintf("calf-cache:%s", c.getHomebrewCachePath())
}

// getNpmCachePath returns the host path for npm cache.
func (c *CacheManager) getNpmCachePath() string {
	return filepath.Join(c.cacheBaseDir, npmCacheDir)
}

// getGoCachePath returns the host path for Go cache.
func (c *CacheManager) getGoCachePath() string {
	return filepath.Join(c.cacheBaseDir, goCacheDir)
}

// SetupHomebrewCache sets up the Homebrew cache directory on the host.
// Creates the cache directory structure with graceful degradation on errors.
func (c *CacheManager) SetupHomebrewCache() error {
	if c.homeDir == "" {
		fmt.Fprintf(os.Stderr, "Warning: home directory not available, continuing without Homebrew cache\n")
		return nil
	}

	hostCacheDir := c.getHomebrewCachePath()

	if err := os.MkdirAll(hostCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create host cache directory: %w", err)
	}

	downloadsDir := filepath.Join(hostCacheDir, homebrewDownloadsDir)
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create downloads directory: %w", err)
	}

	caskDir := filepath.Join(hostCacheDir, homebrewCaskDir)
	if err := os.MkdirAll(caskDir, 0755); err != nil {
		return fmt.Errorf("failed to create Cask directory: %w", err)
	}

	return nil
}

// SetupVMHomebrewCache returns shell commands to set up Homebrew cache in the VM.
// The commands verify that the cache mount exists and is accessible.
// Mount is handled by calf-mount-shares.sh via LaunchDaemon.
// Returns nil if host cache is not available.
func (c *CacheManager) SetupVMHomebrewCache() []string {
	if c.homeDir == "" {
		return nil
	}

	hostCacheDir := c.getHomebrewCachePath()
	if _, err := os.Stat(hostCacheDir); os.IsNotExist(err) {
		return nil
	}

	vmCacheDir := "~/.calf-cache/homebrew"

	commands := []string{
		// Verify mount exists (calf-mount-shares.sh should have created it)
		"mount | grep -q \" on $HOME/.calf-cache \" 2>/dev/null || echo 'Warning: ~/.calf-cache not mounted'",
		// Verify cache subdirectory is accessible
		fmt.Sprintf("test -d %s || echo 'Warning: Homebrew cache directory not found'", vmCacheDir),
		// Configure environment variable
		fmt.Sprintf("touch ~/.zshrc && grep -q 'HOMEBREW_CACHE' ~/.zshrc || echo 'export HOMEBREW_CACHE=%s' >> ~/.zshrc", vmCacheDir),
	}

	return commands
}

// getCacheInfo returns information about a cache at the given path.
// Follows symlinks to report size from actual data location.
func (c *CacheManager) getCacheInfo(cachePath string) (*CacheInfo, error) {
	realPath, err := c.resolveRealCachePath(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve cache path: %w", err)
	}

	pathForSize := cachePath
	if realPath != "" {
		pathForSize = realPath
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &CacheInfo{
				Path:      cachePath,
				Size:      0,
				Available: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to stat cache directory: %w", err)
	}

	return &CacheInfo{
		Path:       cachePath,
		Size:       getDiskUsage(pathForSize),
		Available:  true,
		LastAccess: info.ModTime(),
	}, nil
}

// GetHomebrewCacheInfo returns information about the Homebrew cache.
// Follows symlinks to report size from actual data location.
func (c *CacheManager) GetHomebrewCacheInfo() (*CacheInfo, error) {
	return c.getCacheInfo(c.getHomebrewCachePath())
}

// SetupNpmCache sets up the npm cache directory on the host.
// Creates the cache directory with graceful degradation on errors.
func (c *CacheManager) SetupNpmCache() error {
	if c.homeDir == "" {
		fmt.Fprintf(os.Stderr, "Warning: home directory not available, continuing without npm cache\n")
		return nil
	}

	hostCacheDir := c.getNpmCachePath()

	if err := os.MkdirAll(hostCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create host npm cache directory: %w", err)
	}

	return nil
}

// SetupVMNpmCache returns shell commands to set up npm cache in the VM.
// The commands verify that the cache mount exists and configure npm.
// Mount is handled by calf-mount-shares.sh via LaunchDaemon.
// Returns nil if host cache is not available.
func (c *CacheManager) SetupVMNpmCache() []string {
	if c.homeDir == "" {
		return nil
	}

	hostCacheDir := c.getNpmCachePath()
	if _, err := os.Stat(hostCacheDir); os.IsNotExist(err) {
		return nil
	}

	vmCacheDir := "~/.calf-cache/npm"

	commands := []string{
		// Verify mount exists (calf-mount-shares.sh should have created it)
		"mount | grep -q \" on $HOME/.calf-cache \" 2>/dev/null || echo 'Warning: ~/.calf-cache not mounted'",
		// Verify cache subdirectory is accessible
		fmt.Sprintf("test -d %s || echo 'Warning: npm cache directory not found'", vmCacheDir),
		// Configure npm cache directory
		fmt.Sprintf("npm config set cache %s", vmCacheDir),
	}

	return commands
}

// GetNpmCacheInfo returns information about the npm cache.
// Follows symlinks to report size from actual data location.
func (c *CacheManager) GetNpmCacheInfo() (*CacheInfo, error) {
	return c.getCacheInfo(c.getNpmCachePath())
}

// SetupGoCache sets up the Go cache directory on the host.
// Creates the cache directory structure with graceful degradation on errors.
func (c *CacheManager) SetupGoCache() error {
	if c.homeDir == "" {
		fmt.Fprintf(os.Stderr, "Warning: home directory not available, continuing without Go cache\n")
		return nil
	}

	hostCacheDir := c.getGoCachePath()

	if err := os.MkdirAll(hostCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create host Go cache directory: %w", err)
	}

	pkgModDir := filepath.Join(hostCacheDir, "pkg", "mod")
	if err := os.MkdirAll(pkgModDir, 0755); err != nil {
		return fmt.Errorf("failed to create pkg/mod directory: %w", err)
	}

	pkgSumdbDir := filepath.Join(hostCacheDir, "pkg", "sumdb")
	if err := os.MkdirAll(pkgSumdbDir, 0755); err != nil {
		return fmt.Errorf("failed to create pkg/sumdb directory: %w", err)
	}

	return nil
}

// SetupVMGoCache returns shell commands to set up Go cache in the VM.
// The commands verify that the cache mount exists and configure GOMODCACHE.
// Mount is handled by calf-mount-shares.sh via LaunchDaemon.
// Returns nil if host cache is not available.
func (c *CacheManager) SetupVMGoCache() []string {
	if c.homeDir == "" {
		return nil
	}

	hostCacheDir := c.getGoCachePath()
	if _, err := os.Stat(hostCacheDir); os.IsNotExist(err) {
		return nil
	}

	vmCacheDir := "~/.calf-cache/go"
	gomodcachePath := "~/.calf-cache/go/pkg/mod"

	commands := []string{
		// Verify mount exists (calf-mount-shares.sh should have created it)
		"mount | grep -q \" on $HOME/.calf-cache \" 2>/dev/null || echo 'Warning: ~/.calf-cache not mounted'",
		// Verify cache subdirectory is accessible
		fmt.Sprintf("test -d %s || echo 'Warning: Go cache directory not found'", vmCacheDir),
		// Configure environment variable
		fmt.Sprintf("touch ~/.zshrc && grep -q 'GOMODCACHE' ~/.zshrc || echo 'export GOMODCACHE=%s' >> ~/.zshrc", gomodcachePath),
	}

	return commands
}

// GetGoCacheInfo returns information about the Go cache.
// Follows symlinks to report size from actual data location.
func (c *CacheManager) GetGoCacheInfo() (*CacheInfo, error) {
	return c.getCacheInfo(c.getGoCachePath())
}

// getGitCachePath returns the host path for git cache.
func (c *CacheManager) getGitCachePath() string {
	return filepath.Join(c.cacheBaseDir, gitCacheDir)
}

// resolveRealCachePath resolves the real cache path by following symlinks.
// Always starts from the local cache path (~/.calf-cache/{type}) and follows symlinks if present.
// Returns the real path where data is stored, or empty string if path doesn't exist.
// Only resolves symlinks when using default cache base directory (prevents test interference).
func (c *CacheManager) resolveRealCachePath(localPath string) (string, error) {
	// Only resolve symlinks if using the default cache base directory
	// This prevents tests with temporary directories from accessing shared volumes
	if c.homeDir == "" {
		return "", nil
	}
	expectedCacheBaseDir := filepath.Join(c.homeDir, ".calf-cache")
	if c.cacheBaseDir != expectedCacheBaseDir {
		return "", nil
	}

	// Check if path exists
	info, err := os.Lstat(localPath) // Use Lstat to not follow symlinks yet
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Path doesn't exist
		}
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	// If it's a symlink, resolve it to find where data actually lives
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := filepath.EvalSymlinks(localPath)
		if err != nil {
			// If the symlink exists but its target doesn't exist, treat as unavailable
			// This happens when ~/.calf-cache/{type} is a symlink to a shared volume that isn't mounted
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", fmt.Errorf("failed to resolve symlink: %w", err)
		}
		return target, nil
	}

	// Not a symlink - path doesn't exist yet or is a regular directory
	return "", nil
}

// SetupGitCache sets up the git cache directory on the host.
// Creates the cache directory with graceful degradation on errors.
func (c *CacheManager) SetupGitCache() error {
	if c.homeDir == "" {
		fmt.Fprintf(os.Stderr, "Warning: home directory not available, continuing without git cache\n")
		return nil
	}

	hostCacheDir := c.getGitCachePath()

	if err := os.MkdirAll(hostCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create host git cache directory: %w", err)
	}

	return nil
}

// SetupVMGitCache returns shell commands to set up git cache in the VM.
// The commands verify that the cache mount exists and is accessible.
// Mount is handled by calf-mount-shares.sh via LaunchDaemon.
// Returns nil if host cache is not available.
func (c *CacheManager) SetupVMGitCache() []string {
	if c.homeDir == "" {
		return nil
	}

	hostCacheDir := c.getGitCachePath()
	if _, err := os.Stat(hostCacheDir); os.IsNotExist(err) {
		return nil
	}

	vmCacheDir := "~/.calf-cache/git"

	commands := []string{
		// Verify mount exists (calf-mount-shares.sh should have created it)
		"mount | grep -q \" on $HOME/.calf-cache \" 2>/dev/null || echo 'Warning: ~/.calf-cache not mounted'",
		// Verify cache subdirectory is accessible
		fmt.Sprintf("test -d %s || echo 'Warning: Git cache directory not found'", vmCacheDir),
	}

	return commands
}

// GetGitCacheInfo returns information about the git cache.
// Follows symlinks to report size from actual data location.
func (c *CacheManager) GetGitCacheInfo() (*CacheInfo, error) {
	return c.getCacheInfo(c.getGitCachePath())
}

// GetCachedGitRepos returns a list of git repository names that are cached.
func (c *CacheManager) GetCachedGitRepos() ([]string, error) {
	cachePath := c.getGitCachePath()

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read git cache directory: %w", err)
	}

	repos := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			repos = append(repos, entry.Name())
		}
	}

	return repos, nil
}

// CacheGitRepo clones a git repository to the cache directory.
// repoURL should be the full git URL (e.g., https://github.com/user/repo.git)
// repoName is the name for the cached repo directory (e.g., "repo")
// Returns true if cache was created/updated, false if repo already exists and is up to date.
func (c *CacheManager) CacheGitRepo(repoURL, repoName string) (bool, error) {
	if c.homeDir == "" {
		return false, fmt.Errorf("home directory not available")
	}

	repoCacheDir := filepath.Join(c.getGitCachePath(), repoName)

	if _, err := os.Stat(repoCacheDir); err == nil {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(repoCacheDir), 0755); err != nil {
		return false, fmt.Errorf("failed to create cache directory: %w", err)
	}

	cmd := exec.Command("git", "clone", repoURL, repoCacheDir)
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to clone repo %s: %w", repoURL, err)
	}

	return true, nil
}

// UpdateGitRepos updates all cached git repositories by running git fetch.
// Returns the number of repos updated and any errors encountered.
func (c *CacheManager) UpdateGitRepos() (int, error) {
	repos, err := c.GetCachedGitRepos()
	if err != nil {
		return 0, fmt.Errorf("failed to get cached repos: %w", err)
	}

	if len(repos) == 0 {
		return 0, nil
	}

	updated := 0
	failed := 0
	for _, repo := range repos {
		repoPath := filepath.Join(c.getGitCachePath(), repo)
		if err := exec.Command("git", "-C", repoPath, "rev-parse", "--git-dir").Run(); err != nil {
			// Not a git repository — skip silently
			continue
		}
		if err := exec.Command("git", "-C", repoPath, "fetch", "--all").Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update git cache for %s: %v\n", repo, err)
			failed++
			continue
		}
		updated++
	}

	if failed > 0 {
		return updated, fmt.Errorf("failed to update %d of %d repos", failed, updated+failed)
	}
	return updated, nil
}

// Status displays cache status information to the writer.
func (c *CacheManager) Status(w io.Writer) error {
	homebrewInfo, err := c.GetHomebrewCacheInfo()
	if err != nil {
		return fmt.Errorf("failed to get Homebrew cache info: %w", err)
	}

	npmInfo, err := c.GetNpmCacheInfo()
	if err != nil {
		return fmt.Errorf("failed to get npm cache info: %w", err)
	}

	goInfo, err := c.GetGoCacheInfo()
	if err != nil {
		return fmt.Errorf("failed to get Go cache info: %w", err)
	}

	gitInfo, err := c.GetGitCacheInfo()
	if err != nil {
		return fmt.Errorf("failed to get git cache info: %w", err)
	}

	fmt.Fprintf(w, "Cache Status:\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Homebrew:\n")
	fmt.Fprintf(w, "  Location: %s\n", homebrewInfo.Path)
	fmt.Fprintf(w, "  Status: ")
	if homebrewInfo.Available {
		fmt.Fprintf(w, "✓ Ready\n")
		fmt.Fprintf(w, "  Size: %s\n", FormatBytes(homebrewInfo.Size))
		if !homebrewInfo.LastAccess.IsZero() {
			fmt.Fprintf(w, "  Last access: %s\n", homebrewInfo.LastAccess.Format(time.RFC3339))
		}
	} else {
		fmt.Fprintf(w, "✗ Not configured\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "npm:\n")
	fmt.Fprintf(w, "  Location: %s\n", npmInfo.Path)
	fmt.Fprintf(w, "  Status: ")
	if npmInfo.Available {
		fmt.Fprintf(w, "✓ Ready\n")
		fmt.Fprintf(w, "  Size: %s\n", FormatBytes(npmInfo.Size))
		if !npmInfo.LastAccess.IsZero() {
			fmt.Fprintf(w, "  Last access: %s\n", npmInfo.LastAccess.Format(time.RFC3339))
		}
	} else {
		fmt.Fprintf(w, "✗ Not configured\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Go:\n")
	fmt.Fprintf(w, "  Location: %s\n", goInfo.Path)
	fmt.Fprintf(w, "  Status: ")
	if goInfo.Available {
		fmt.Fprintf(w, "✓ Ready\n")
		fmt.Fprintf(w, "  Size: %s\n", FormatBytes(goInfo.Size))
		if !goInfo.LastAccess.IsZero() {
			fmt.Fprintf(w, "  Last access: %s\n", goInfo.LastAccess.Format(time.RFC3339))
		}
	} else {
		fmt.Fprintf(w, "✗ Not configured\n")
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "Git:\n")
	fmt.Fprintf(w, "  Location: %s\n", gitInfo.Path)
	fmt.Fprintf(w, "  Status: ")
	if gitInfo.Available {
		fmt.Fprintf(w, "✓ Ready\n")
		fmt.Fprintf(w, "  Size: %s\n", FormatBytes(gitInfo.Size))
		if !gitInfo.LastAccess.IsZero() {
			fmt.Fprintf(w, "  Last access: %s\n", gitInfo.LastAccess.Format(time.RFC3339))
		}
		repos, err := c.GetCachedGitRepos()
		if err == nil && len(repos) > 0 {
			fmt.Fprintf(w, "  Cached repos: %d\n", len(repos))
			for _, repo := range repos {
				fmt.Fprintf(w, "    - %s\n", repo)
			}
		}
	} else {
		fmt.Fprintf(w, "✗ Not configured\n")
	}
	fmt.Fprintf(w, "\n")

	return nil
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// clearDirectoryContents removes all contents of a directory without removing the directory itself.
// Used when clearing symlinked caches to preserve the symlink.
func clearDirectoryContents(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(dirPath, entry.Name())
		if err := removeAllWithPermFix(entryPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// removeAllWithPermFix removes a directory tree, fixing permissions as needed.
// Go module cache files may have read-only permissions that prevent deletion.
// This function makes all files and directories writable before attempting to delete.
func removeAllWithPermFix(path string) error {
	// First, try the fast path - works for most caches
	err := os.RemoveAll(path)
	if err == nil {
		return nil
	}

	// If we got a permission error, walk the tree and fix permissions
	if os.IsPermission(err) {
		// Make everything in the tree writable
		filepath.Walk(path, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				// Ignore walk errors - we'll try to delete what we can
				return nil
			}
			// Make writable: add owner write permission (0200)
			// Ignore chmod errors - we'll try to delete anyway
			os.Chmod(p, info.Mode()|0200)
			return nil
		})

		// Try removing again after fixing permissions
		err = os.RemoveAll(path)
	}

	return err
}

// Clear removes the specified cache type and recreates an empty cache directory.
// cacheType must be one of: "homebrew", "npm", "go", "git"
// dryRun if true, simulates clearing without actually deleting files
// Returns true if cache was cleared (or would be cleared in dry run), false if cache didn't exist
// Always starts at ~/.calf-cache/{type} and follows symlinks to clear actual data.
// If the cache path is a symlink (e.g., in a VM), preserves the symlink and only clears contents.
func (c *CacheManager) Clear(cacheType string, dryRun bool) (bool, error) {
	if c.homeDir == "" {
		return false, fmt.Errorf("home directory not available")
	}

	var localCachePath string
	var setupFunc func() error

	switch cacheType {
	case "homebrew":
		localCachePath = c.getHomebrewCachePath()
		setupFunc = c.SetupHomebrewCache
	case "npm":
		localCachePath = c.getNpmCachePath()
		setupFunc = c.SetupNpmCache
	case "go":
		localCachePath = c.getGoCachePath()
		setupFunc = c.SetupGoCache
	case "git":
		localCachePath = c.getGitCachePath()
		setupFunc = c.SetupGitCache
	default:
		return false, fmt.Errorf("invalid cache type: %s (must be homebrew, npm, go, or git)", cacheType)
	}

	// Check if local path exists
	info, err := os.Lstat(localCachePath) // Use Lstat to detect symlinks
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // Cache doesn't exist
		}
		return false, fmt.Errorf("failed to check cache directory: %w", err)
	}

	// Check if this is a symlink (e.g., inside a VM pointing to shared volume)
	isSymlink := info.Mode()&os.ModeSymlink != 0

	// Resolve symlink to find where data actually lives
	realPath, err := c.resolveRealCachePath(localCachePath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve cache path: %w", err)
	}

	if !dryRun {
		if isSymlink && realPath != "" {
			// For symlinked caches (VM scenario):
			// - Clear contents of the target directory (shared volume)
			// - Preserve the symlink itself so cache continues working after clear
			if err := clearDirectoryContents(realPath); err != nil {
				return false, fmt.Errorf("failed to clear cache contents: %w", err)
			}
		} else {
			// For regular directory caches (host scenario):
			// - Remove the entire directory
			// - Recreate empty structure
			if err := removeAllWithPermFix(localCachePath); err != nil {
				return false, fmt.Errorf("failed to remove cache directory: %w", err)
			}

			if err := setupFunc(); err != nil {
				return false, fmt.Errorf("failed to recreate cache directory: %w", err)
			}
		}
	}

	return true, nil
}
