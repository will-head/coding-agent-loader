package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading config when file doesn't exist (should use defaults)
	t.Run("when config file is missing should use default values", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Act
		cfg, err := LoadConfig(configPath, "")

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 4 {
			t.Errorf("Expected CPU default 4, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 8192 {
			t.Errorf("Expected Memory default 8192, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
		if cfg.Isolation.Defaults.VM.DiskSize != 80 {
			t.Errorf("Expected DiskSize default 80, got %d", cfg.Isolation.Defaults.VM.DiskSize)
		}
		if cfg.Isolation.Defaults.VM.BaseImage != "ghcr.io/cirruslabs/macos-sequoia-base:latest" {
			t.Errorf("Expected BaseImage default 'ghcr.io/cirruslabs/macos-sequoia-base:latest', got %s", cfg.Isolation.Defaults.VM.BaseImage)
		}
		if cfg.Isolation.Defaults.Proxy.Mode != "auto" {
			t.Errorf("Expected Proxy mode default 'auto', got %s", cfg.Isolation.Defaults.Proxy.Mode)
		}
	})

	// Test loading config when both paths are empty (should use defaults)
	t.Run("when both paths are empty should use default values", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		cfg, err := LoadConfig("", "")

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig with empty paths returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 4 {
			t.Errorf("Expected CPU default 4, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 8192 {
			t.Errorf("Expected Memory default 8192, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
	})

	// Test loading valid config file
	t.Run("when valid config file exists should load all fields correctly", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: 8
      memory: 16384
      disk_size: 120
      base_image: "custom-image:latest"
    github:
      default_branch_prefix: "feature/"
    output:
      sync_dir: "~/my-output"
    proxy:
      mode: "on"
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Act
		cfg, err := LoadConfig(configPath, "")

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 8 {
			t.Errorf("Expected CPU 8, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 16384 {
			t.Errorf("Expected Memory 16384, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
		if cfg.Isolation.Defaults.VM.DiskSize != 120 {
			t.Errorf("Expected DiskSize 120, got %d", cfg.Isolation.Defaults.VM.DiskSize)
		}
		if cfg.Isolation.Defaults.VM.BaseImage != "custom-image:latest" {
			t.Errorf("Expected BaseImage 'custom-image:latest', got %s", cfg.Isolation.Defaults.VM.BaseImage)
		}
		if cfg.Isolation.Defaults.GitHub.DefaultBranchPrefix != "feature/" {
			t.Errorf("Expected DefaultBranchPrefix 'feature/', got %s", cfg.Isolation.Defaults.GitHub.DefaultBranchPrefix)
		}
		if cfg.Isolation.Defaults.Output.SyncDir != "~/my-output" {
			t.Errorf("Expected SyncDir '~/my-output', got %s", cfg.Isolation.Defaults.Output.SyncDir)
		}
		if cfg.Isolation.Defaults.Proxy.Mode != "on" {
			t.Errorf("Expected Proxy mode 'on', got %s", cfg.Isolation.Defaults.Proxy.Mode)
		}
	})

	// Test partial config file (some fields missing)
	t.Run("when config file has only some fields should use defaults for missing fields", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: 8
      memory: 16384
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Act
		cfg, err := LoadConfig(configPath, "")

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 8 {
			t.Errorf("Expected CPU 8, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 16384 {
			t.Errorf("Expected Memory 16384, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
		// Verify defaults for missing fields
		if cfg.Isolation.Defaults.VM.DiskSize != 80 {
			t.Errorf("Expected DiskSize default 80, got %d", cfg.Isolation.Defaults.VM.DiskSize)
		}
		if cfg.Isolation.Defaults.VM.BaseImage != "ghcr.io/cirruslabs/macos-sequoia-base:latest" {
			t.Errorf("Expected BaseImage default, got %s", cfg.Isolation.Defaults.VM.BaseImage)
		}
	})

	// Test malformed YAML file (should return error)
	t.Run("when config file contains malformed YAML should return parse error", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		malformedContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: [this is not valid yaml syntax
`
		if err := os.WriteFile(configPath, []byte(malformedContent), 0644); err != nil {
			t.Fatalf("Failed to write malformed config: %v", err)
		}

		// Act
		_, err := LoadConfig(configPath, "")

		// Assert
		if err == nil {
			t.Error("Expected error for malformed YAML, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse config file") {
			t.Errorf("Expected parse error message, got: %v", err)
		}
	})
}

func TestLoadVMConfig(t *testing.T) {
	// Test loading per-VM config
	t.Run("when vm config file exists should override global config fields", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		globalConfigPath := filepath.Join(tmpDir, "config.yaml")
		vmDir := filepath.Join(tmpDir, "vms", "test-vm")
		vmConfigPath := filepath.Join(vmDir, "vm.yaml")
		globalConfigContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: 4
      memory: 8192
      disk_size: 80
`
		if err := os.WriteFile(globalConfigPath, []byte(globalConfigContent), 0644); err != nil {
			t.Fatalf("Failed to write global config: %v", err)
		}
		if err := os.MkdirAll(vmDir, 0755); err != nil {
			t.Fatalf("Failed to create VM directory: %v", err)
		}
		vmConfigContent := `
cpu: 8
memory: 16384
`
		if err := os.WriteFile(vmConfigPath, []byte(vmConfigContent), 0644); err != nil {
			t.Fatalf("Failed to write VM config: %v", err)
		}

		// Act
		cfg, err := LoadConfig(globalConfigPath, vmConfigPath)

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 8 {
			t.Errorf("Expected CPU 8 from VM config, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 16384 {
			t.Errorf("Expected Memory 16384 from VM config, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
		// Verify global default for missing field
		if cfg.Isolation.Defaults.VM.DiskSize != 80 {
			t.Errorf("Expected DiskSize 80 from global config, got %d", cfg.Isolation.Defaults.VM.DiskSize)
		}
	})

	// Test missing per-VM config (should use global/defaults)
	t.Run("when vm config file is missing should use global config values", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		globalConfigPath := filepath.Join(tmpDir, "config.yaml")
		globalConfigContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: 8
      memory: 16384
`
		if err := os.WriteFile(globalConfigPath, []byte(globalConfigContent), 0644); err != nil {
			t.Fatalf("Failed to write global config: %v", err)
		}

		// Act
		cfg, err := LoadConfig(globalConfigPath, "")

		// Assert
		if err != nil {
			t.Fatalf("LoadConfig returned unexpected error: %v", err)
		}
		if cfg.Isolation.Defaults.VM.CPU != 8 {
			t.Errorf("Expected CPU 8, got %d", cfg.Isolation.Defaults.VM.CPU)
		}
		if cfg.Isolation.Defaults.VM.Memory != 16384 {
			t.Errorf("Expected Memory 16384, got %d", cfg.Isolation.Defaults.VM.Memory)
		}
	})

	t.Run("when config file has zero values should fail validation", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
version: 1
isolation:
  defaults:
    vm:
      cpu: 0
      memory: 0
      disk_size: 0
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Act
		_, err := LoadConfig(configPath, "")

		// Assert
		if err == nil {
			t.Error("Expected validation error for zero values in config, got nil")
		}
		expectedMsg := "invalid CPU '0'"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("when config file has empty string fields should fail validation", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := `
version: 1
isolation:
  defaults:
    vm:
      base_image: ""
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Act
		_, err := LoadConfig(configPath, "")

		// Assert
		if err == nil {
			t.Error("Expected validation error for empty base image in config, got nil")
		}
		expectedMsg := "invalid base_image ''"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestValidateConfig(t *testing.T) {
	t.Run("when all fields are valid should pass validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:       4,
						Memory:    8192,
						DiskSize:  80,
						BaseImage: "ghcr.io/cirruslabs/macos-sequoia-base:latest",
					},
					Proxy: ProxyConfig{
						Mode: "auto",
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err != nil {
			t.Errorf("Validate returned unexpected error: %v", err)
		}
	})

	t.Run("when cpu is zero should fail validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU: 0, // Invalid
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err == nil {
			t.Error("Expected validation error for invalid CPU, got nil")
		}
		expectedMsg := "invalid CPU '0'"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("when memory is below minimum should fail validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:    4,   // Valid
						Memory: 100, // Invalid (below 256 MB minimum)
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err == nil {
			t.Error("Expected validation error for invalid memory, got nil")
		}
		expectedMsg := "invalid memory '100'"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("when memory is at minimum threshold should pass validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:       4,
						Memory:    256, // Tart minimum (v2.4.3+)
						DiskSize:  80,
						BaseImage: "test-image",
					},
					Proxy: ProxyConfig{
						Mode: "auto",
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err != nil {
			t.Errorf("Validate returned unexpected error for minimum valid memory: %v", err)
		}
	})

	t.Run("when disk size is zero should fail validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:      4,    // Valid
						Memory:   8192, // Valid
						DiskSize: 0,    // Invalid
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err == nil {
			t.Error("Expected validation error for invalid disk size, got nil")
		}
		expectedMsg := "invalid disk_size '0'"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("when proxy mode is unrecognised should fail validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:       4,                                              // Valid
						Memory:    8192,                                           // Valid
						DiskSize:  80,                                             // Valid
						BaseImage: "ghcr.io/cirruslabs/macos-sequoia-base:latest", // Valid
					},
					Proxy: ProxyConfig{
						Mode: "invalid", // Invalid
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err == nil {
			t.Error("Expected validation error for invalid proxy mode, got nil")
		}
		expectedMsg := "invalid proxy mode 'invalid'"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("when base image is empty should fail validation", func(t *testing.T) {
		// Arrange
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU:       4,
						Memory:    8192,
						DiskSize:  80,
						BaseImage: "", // Invalid
					},
				},
			},
		}

		// Act
		err := cfg.Validate("")

		// Assert
		if err == nil {
			t.Error("Expected validation error for empty base image, got nil")
		}
		expectedMsg := "invalid base_image ''"
		if !strings.HasPrefix(err.Error(), expectedMsg) {
			t.Errorf("Expected error message to start with '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestGetDefaultConfigPath(t *testing.T) {
	t.Run("when home dir is available should return path ending in .calf/config.yaml", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		path, err := GetDefaultConfigPath()

		// Assert
		if err != nil {
			t.Fatalf("GetDefaultConfigPath returned unexpected error: %v", err)
		}
		if !strings.HasSuffix(path, filepath.Join(".calf", "config.yaml")) {
			t.Errorf("Expected path ending in .calf/config.yaml, got %s", path)
		}
	})

	t.Run("when home dir is available should return an absolute path", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		path, err := GetDefaultConfigPath()

		// Assert
		if err != nil {
			t.Fatalf("GetDefaultConfigPath returned unexpected error: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got %s", path)
		}
	})
}

func TestGetVMConfigPath(t *testing.T) {
	t.Run("when vm name is provided should return path containing the vm name", func(t *testing.T) {
		// Arrange
		vmName := "calf-dev"

		// Act
		path, err := GetVMConfigPath(vmName)

		// Assert
		if err != nil {
			t.Fatalf("GetVMConfigPath returned unexpected error: %v", err)
		}
		if !strings.Contains(path, vmName) {
			t.Errorf("Expected path to contain vm name %q, got %s", vmName, path)
		}
	})

	t.Run("when vm name is provided should return path ending in vm.yaml", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		path, err := GetVMConfigPath("calf-dev")

		// Assert
		if err != nil {
			t.Fatalf("GetVMConfigPath returned unexpected error: %v", err)
		}
		if !strings.HasSuffix(path, "vm.yaml") {
			t.Errorf("Expected path ending in vm.yaml, got %s", path)
		}
	})

	t.Run("when vm name is provided should return an absolute path", func(t *testing.T) {
		// Arrange — no setup needed

		// Act
		path, err := GetVMConfigPath("calf-dev")

		// Assert
		if err != nil {
			t.Fatalf("GetVMConfigPath returned unexpected error: %v", err)
		}
		if !filepath.IsAbs(path) {
			t.Errorf("Expected absolute path, got %s", path)
		}
	})
}

func TestConfigPathValidation(t *testing.T) {
	t.Run("when global config has validation error should include file path in error message", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU: 0,
					},
				},
			},
		}

		// Act
		err := cfg.Validate(configPath)

		// Assert
		if err == nil {
			t.Error("Expected validation error, got nil")
		}
		if !strings.Contains(err.Error(), configPath) {
			t.Errorf("Expected error message to include config path '%s', got '%s'", configPath, err.Error())
		}
	})

	t.Run("when vm config has validation error should include file path in error message", func(t *testing.T) {
		// Arrange
		tmpDir := t.TempDir()
		vmConfigPath := filepath.Join(tmpDir, "vm.yaml")
		cfg := &Config{
			Version: 1,
			Isolation: IsolationConfig{
				Defaults: DefaultsConfig{
					VM: VMConfig{
						CPU: 0,
					},
				},
			},
		}

		// Act
		err := cfg.Validate(vmConfigPath)

		// Assert
		if err == nil {
			t.Error("Expected validation error, got nil")
		}
		if !strings.Contains(err.Error(), vmConfigPath) {
			t.Errorf("Expected error message to include VM config path '%s', got '%s'", vmConfigPath, err.Error())
		}
	})
}
