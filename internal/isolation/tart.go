// Package isolation provides VM isolation and management for CALF.
package isolation

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	// TartInstallPrompt is the message shown when Tart is not installed.
	TartInstallPrompt = "Tart is not installed. Install via Homebrew? [Y/n]: "

	// Default IP polling interval.
	defaultPollInterval = 2 * time.Second

	// Default IP polling timeout.
	defaultPollTimeout = 60 * time.Second

	// Cache sharing directory mount path (read-only).
	// TODO: Use virtio-fs mount system (like calf-cache) instead of symlink approach.
	// calf-bootstrap uses symlinks for tart-cache (fine for bootstrap), but the Go
	// implementation should mount /Volumes/My Shared Files/tart-cache via virtio-fs
	// for consistency with the calf-cache architecture. See calf-mount-shares.sh and
	// ADR-003 for mount system details.
	cacheDirMount = "tart-cache:~/.tart/cache:ro"
)

// VMState represents the current state of a Tart VM.
type VMState string

const (
	// StateRunning indicates the VM is currently running.
	StateRunning VMState = "running"

	// StateStopped indicates the VM is stopped but exists.
	StateStopped VMState = "stopped"

	// StateNotFound indicates the VM does not exist.
	StateNotFound VMState = "not_found"
)

// VMInfo contains information about a Tart VM.
type VMInfo struct {
	Name  string  `json:"name"`
	State VMState `json:"state"`
	Size  float64 `json:"size,omitempty"`
}

// TartListOutput is the JSON output from `tart list --format json`.
type TartListOutput []VMInfo

// commandRunner is a function type for executing commands (allows mocking in tests).
type commandRunner func(args ...string) (string, error)

// TartClient wraps the Tart CLI for VM operations.
type TartClient struct {
	tartPath       string
	installPrompt  string
	outputWriter   io.Writer
	errorWriter    io.Writer
	pollInterval   time.Duration
	pollTimeout    time.Duration
	runCommand     commandRunner
	runBrewCommand commandRunner
	stdinReader    io.Reader
	lookPath       func(string) (string, error)
}

// NewTartClient creates a new TartClient.
func NewTartClient() *TartClient {
	client := &TartClient{
		installPrompt: TartInstallPrompt,
		outputWriter:  os.Stdout,
		errorWriter:   os.Stderr,
		pollInterval:  defaultPollInterval,
		pollTimeout:   defaultPollTimeout,
		stdinReader:   os.Stdin,
		lookPath:      exec.LookPath,
	}
	// Set default command runners
	client.runCommand = client.runTartCommand
	client.runBrewCommand = func(args ...string) (string, error) {
		brewPath, err := client.lookPath("brew")
		if err != nil {
			return "", fmt.Errorf("brew not found: %w", err)
		}
		cmd := exec.Command(brewPath, args...)
		cmd.Stdout = client.outputWriter
		cmd.Stderr = client.errorWriter
		if err := cmd.Run(); err != nil {
			return "", err
		}
		return "", nil
	}
	return client
}

// ensureInstalled checks if Tart is installed and offers to install via Homebrew if not.
func (c *TartClient) ensureInstalled() error {
	if c.tartPath != "" {
		return nil
	}

	path, err := c.lookPath("tart")
	if err == nil {
		c.tartPath = path
		return nil
	}

	if _, err := c.lookPath("brew"); err != nil {
		return fmt.Errorf("tart is not installed and Homebrew is not available. Please install Tart manually: https://github.com/cirruslabs/tart")
	}

	fmt.Fprint(c.errorWriter, c.installPrompt)

	reader := bufio.NewReader(c.stdinReader)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" && response != "y" && response != "yes" {
		return fmt.Errorf("tart installation cancelled")
	}

	fmt.Fprintln(c.outputWriter, "Installing Tart via Homebrew...")
	if _, err := c.runBrewCommand("install", "cirruslabs/cli/tart"); err != nil {
		return fmt.Errorf("failed to install Tart: %w", err)
	}

	path, err = c.lookPath("tart")
	if err != nil {
		return fmt.Errorf("tart installation completed but 'tart' command not found in PATH")
	}

	c.tartPath = path
	fmt.Fprintln(c.outputWriter, "tart installed successfully")
	return nil
}

// runTartCommand executes a Tart CLI command and returns combined stdout/stderr.
func (c *TartClient) runTartCommand(args ...string) (string, error) {
	cmd := exec.Command(c.tartPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("tart %s failed: %w\nstdout: %s\nstderr: %s",
			strings.Join(args, " "), err, stdout.String(), stderr.String())
	}

	return stdout.String(), nil
}

// Clone clones a VM from an image or local VM.
func (c *TartClient) Clone(image, name string) error {
	if err := c.ensureInstalled(); err != nil {
		return err
	}
	if _, err := c.runCommand("clone", image, name); err != nil {
		return fmt.Errorf("failed to clone VM %s from %s: %w", name, image, err)
	}
	return nil
}

// Set configures VM resources (CPU, memory, disk size).
func (c *TartClient) Set(name string, cpu int, memory int, disk string) error {
	if err := c.ensureInstalled(); err != nil {
		return err
	}
	args := []string{"set", name}

	if cpu > 0 {
		args = append(args, fmt.Sprintf("--cpu=%d", cpu))
	}

	if memory > 0 {
		args = append(args, fmt.Sprintf("--memory=%d", memory))
	}

	if disk != "" {
		args = append(args, fmt.Sprintf("--disk-size=%s", disk))
	}

	if _, err := c.runCommand(args...); err != nil {
		return fmt.Errorf("failed to configure VM %s: %w", name, err)
	}

	return nil
}

// Run starts a VM with optional headless mode, VNC, and directory sharing.
// When vnc is true, always uses --vnc-experimental for bidirectional clipboard.
func (c *TartClient) Run(name string, headless, vnc bool, dirs []string) error {
	return c.RunWithCacheDirs(name, headless, vnc, dirs, []string{})
}

// RunWithCacheDirs starts a VM with cache directories shared.
// cacheDirs specifies additional directories to share for caching (e.g., Homebrew cache).
func (c *TartClient) RunWithCacheDirs(name string, headless, vnc bool, dirs []string, cacheDirs []string) error {
	if err := c.ensureInstalled(); err != nil {
		return err
	}
	args := []string{"run"}

	if headless {
		args = append(args, "--headless")
	}

	if vnc {
		args = append(args, "--vnc-experimental")
	}

	args = append(args, fmt.Sprintf("--dir=%s", cacheDirMount))

	for _, dir := range cacheDirs {
		args = append(args, fmt.Sprintf("--dir=%s", dir))
	}

	for _, dir := range dirs {
		args = append(args, fmt.Sprintf("--dir=%s", dir))
	}

	args = append(args, name)

	if _, err := c.runCommand(args...); err != nil {
		return fmt.Errorf("failed to start VM %s: %w", name, err)
	}

	return nil
}

// Stop stops a running VM.
func (c *TartClient) Stop(name string, force bool) error {
	if err := c.ensureInstalled(); err != nil {
		return err
	}
	args := []string{"stop", name}
	if force {
		args = append(args, "--timeout=0")
	}

	if _, err := c.runCommand(args...); err != nil {
		return fmt.Errorf("failed to stop VM %s: %w", name, err)
	}

	return nil
}

// Delete deletes a VM.
func (c *TartClient) Delete(name string) error {
	if err := c.ensureInstalled(); err != nil {
		return err
	}
	if _, err := c.runCommand("delete", name); err != nil {
		return fmt.Errorf("failed to delete VM %s: %w", name, err)
	}
	return nil
}

// List lists all VMs with JSON format for sizes.
func (c *TartClient) List() (TartListOutput, error) {
	if err := c.ensureInstalled(); err != nil {
		return nil, err
	}
	output, err := c.runCommand("list", "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}

	var vms TartListOutput
	if err := json.Unmarshal([]byte(output), &vms); err != nil {
		return nil, fmt.Errorf("failed to parse VM list JSON: %w", err)
	}

	return vms, nil
}

// IP gets the IP address of a running VM, with optional polling for boot.
func (c *TartClient) IP(name string, timeout time.Duration) (string, error) {
	if err := c.ensureInstalled(); err != nil {
		return "", err
	}
	if timeout == 0 {
		timeout = c.pollTimeout
	}

	startTime := time.Now()
	elapsed := 0 * time.Second

	for time.Since(startTime) < timeout {
		output, err := c.runCommand("ip", name)
		if err == nil {
			ip := strings.TrimSpace(output)
			if ip != "" {
				return ip, nil
			}
		}

		elapsed = time.Since(startTime)
		fmt.Fprintf(c.outputWriter, "\rWaiting for VM to boot... %ds", int(elapsed.Seconds()))

		time.Sleep(c.pollInterval)
	}

	fmt.Fprint(c.outputWriter, "\n")
	return "", fmt.Errorf("VM %s did not acquire an IP address within %v", name, timeout)
}

// Get retrieves information about a specific VM.
func (c *TartClient) Get(name string) (*VMInfo, error) {
	vms, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == name {
			return &vm, nil
		}
	}

	return nil, fmt.Errorf("VM %s not found", name)
}

// IsRunning checks if a VM is currently running.
func (c *TartClient) IsRunning(name string) bool {
	state := c.GetState(name)
	return state == StateRunning
}

// Exists checks if a VM exists.
func (c *TartClient) Exists(name string) bool {
	state := c.GetState(name)
	return state != StateNotFound
}

// GetState returns the current state of a VM.
func (c *TartClient) GetState(name string) VMState {
	vm, err := c.Get(name)
	if err != nil {
		return StateNotFound
	}
	return vm.State
}
