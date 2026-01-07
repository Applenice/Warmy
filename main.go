package main

import (
	"fmt"
	"os"
	"path/filepath"

	"warmy/internal/config"
	"warmy/internal/git"
	"warmy/internal/logger"
)

func main() {
	// Parse command line arguments
	if err := parseArgs(); err != nil {
		if err.Error() == "show_help" {
			printHelp()
			return
		}
		if err.Error() == "show_version" {
			printVersion()
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load configuration file
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.InitLogger(cfg.LogLevel)
	log := logger.GetLogger()

	log.WithFields(logger.Fields{
		"program":     "warmy",
		"author":      "https://github.com/Applenice",
		"version":     "1.0.0",
		"config_file": cfg.ConfigFile,
	}).Info("Program started")

	// Get specified commit information
	commitInfo, err := git.GetCommit(cfg.RepoPath, cfg.CommitHash)
	if err != nil {
		log.WithFields(logger.Fields{
			"repo_path":   cfg.RepoPath,
			"commit_hash": cfg.CommitHash,
			"error":       err.Error(),
		}).Fatal("Failed to get commit")
	}

	// Format as JSON
	jsonOutput, err := commitInfo.ToJSON(cfg.PrettyJSON)
	if err != nil {
		log.WithError(err).Fatal("Failed to format JSON")
	}

	// Build output filename
	shortHash := commitInfo.ShortHash
	analyzeTime := commitInfo.AnalyzeTime
	outputFilename := fmt.Sprintf("%s-%s.json", shortHash, analyzeTime)

	// Save output file path to commitInfo
	commitInfo.OutputFile = outputFilename

	// Reformat as JSON (including output file path)
	jsonOutput, err = commitInfo.ToJSON(cfg.PrettyJSON)
	if err != nil {
		log.WithError(err).Fatal("Failed to reformat JSON")
	}

	// Output result to console
	if !cfg.NoConsole {
		fmt.Println(jsonOutput)
		log.Info("JSON data output to console")
	}

	// Save result to file
	if !cfg.NoFile {
		err := saveJSONToFile(cfg.OutputDir, outputFilename, jsonOutput)
		if err != nil {
			log.WithError(err).Error("Failed to save to file")
		} else {
			fullPath := outputFilename
			if cfg.OutputDir != "." && cfg.OutputDir != "" {
				fullPath = filepath.Join(cfg.OutputDir, outputFilename)
			}
			log.WithFields(logger.Fields{
				"filename": outputFilename,
				"filepath": fullPath,
			}).Info("JSON data saved to file")
		}
	}

	log.Info("Program execution completed")
}

// parseArgs parses command line arguments
func parseArgs() error {
	// Only parse --config parameter
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-h", "--help":
			return fmt.Errorf("show_help")
		case "-v", "--version":
			return fmt.Errorf("show_version")
		case "--config":
			if i+1 < len(os.Args) {
				config.SetConfigFile(os.Args[i+1])
				i++
			} else {
				return fmt.Errorf("--config parameter requires specifying config file path")
			}
		}
	}
	return nil
}

// saveJSONToFile saves JSON to file
func saveJSONToFile(dir, filename, data string) error {
	// Ensure directory exists
	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Build complete file path
	filepath := filename
	if dir != "." && dir != "" {
		filepath = dir + "/" + filename
	}

	err := os.WriteFile(filepath, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("failed to save to file: %w", err)
	}

	return nil
}

// printHelp prints help information
func printHelp() {
	helpText := `Warmy Git Commit Reader v1.0.0

Author: https://github.com/Applenice

A configuration file driven Git commit analysis tool with focus feature support.

Usage:
  warmy [options]

Options:
  -h, --help        Show help information
  -v, --version     Show version information
  --config FILE     Specify configuration file path (optional, defaults to config.json in current directory)

Configuration file:
  The program will look for config.json configuration file in the current directory.
  See README.md for detailed configuration documentation.

Examples:
  # Use config.json configuration file in current directory
  warmy
  
  # Specify configuration file
  warmy --config config.json
  
  # Show help
  warmy --help
  
  # Show version
  warmy --version
`
	fmt.Print(helpText)
}

// printVersion prints version information
func printVersion() {
	versionText := "Warmy Git Commit Reader v1.0.0\n"
	fmt.Print(versionText)
}
