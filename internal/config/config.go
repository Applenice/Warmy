package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// FocusConfig focus configuration
type FocusConfig struct {
	Enable         bool     `json:"enable,omitempty"`          // Whether to enable focus
	AddFiles       bool     `json:"add_files,omitempty"`       // Whether to focus on new files
	ModifyFiles    bool     `json:"modify_files,omitempty"`    // Whether to focus on modified files
	DeleteFiles    bool     `json:"delete_files,omitempty"`    // Whether to focus on deleted files
	FilePatterns   []string `json:"file_patterns,omitempty"`   // File path matching patterns
	IgnorePatterns []string `json:"ignore_patterns,omitempty"` // Ignore patterns
}

// Config configuration parameters
type Config struct {
	RepoPath        string      `json:"repo_path,omitempty"`
	CommitHash      string      `json:"commit_hash,omitempty"` // Specify commit hash
	OutputFormat    string      `json:"output_format,omitempty"`
	PrettyJSON      bool        `json:"pretty_json,omitempty"`
	MaxDiffSize     int         `json:"max_diff_size,omitempty"`
	IncludeFullDiff bool        `json:"include_full_diff,omitempty"`
	Verbose         bool        `json:"verbose,omitempty"`
	ParseDiff       bool        `json:"parse_diff,omitempty"`  // Whether to parse diff content
	OutputDir       string      `json:"output_dir,omitempty"`  // Output directory
	NoFile          bool        `json:"no_file,omitempty"`     // Do not output to file
	NoConsole       bool        `json:"no_console,omitempty"`  // Do not output to console
	LogLevel        string      `json:"log_level,omitempty"`   // Log level
	ConfigFile      string      `json:"config_file,omitempty"` // Config file path
	Focus           FocusConfig `json:"focus,omitempty"`       // Focus configuration
}

// Global configuration variable
var globalConfig = Config{
	MaxDiffSize:     1024 * 1024, // Default 1MB
	IncludeFullDiff: false,
	PrettyJSON:      true,
	Verbose:         false,
	ParseDiff:       true, // Default parse diff
	OutputDir:       ".",  // Default current directory
	NoFile:          false,
	NoConsole:       false,
	LogLevel:        "info", // Default log level
	ConfigFile:      "",     // Default no config file
	Focus: FocusConfig{
		Enable:      true,
		AddFiles:    true,
		ModifyFiles: true,
		DeleteFiles: true, // Add delete files focus
		// FilePatterns and IgnorePatterns are now empty by default
		// They must be provided in the config file if focus is enabled
	},
}

// SetConfigFile sets config file path
func SetConfigFile(filename string) {
	globalConfig.ConfigFile = filename
}

// GetConfig gets current configuration
func GetConfig() *Config {
	return &globalConfig
}

// LoadConfig loads configuration from file
func LoadConfig() (*Config, error) {
	// Find config file
	configFile, err := findConfigFile()
	if err != nil {
		return nil, err
	}

	// Load config from file
	fileConfig, err := loadConfigFromFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Update global config
	globalConfig = *fileConfig
	globalConfig.ConfigFile = configFile

	return &globalConfig, nil
}

// findConfigFile finds config file
func findConfigFile() (string, error) {
	// If command line specifies config file, return directly
	if globalConfig.ConfigFile != "" {
		if _, err := os.Stat(globalConfig.ConfigFile); err == nil {
			return globalConfig.ConfigFile, nil
		}
		return "", fmt.Errorf("specified config file does not exist: %s", globalConfig.ConfigFile)
	}

	// Only look for config.json in current directory
	configFile := "config.json"
	if _, err := os.Stat(configFile); err == nil {
		return configFile, nil
	}

	return "", fmt.Errorf("config file not found: %s", configFile)
}

// loadConfigFromFile loads configuration from file
func loadConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
