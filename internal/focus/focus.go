package focus

import (
	"fmt"
	"regexp"

	"warmy/internal/config"
	"warmy/internal/logger"
	"warmy/internal/types"
)

// CompiledPatterns compiled regular expressions
type CompiledPatterns struct {
	FilePatterns   []*regexp.Regexp
	IgnorePatterns []*regexp.Regexp
}

var (
	compiledPatterns *CompiledPatterns
	log              logger.Logger
)

// Init initializes focus feature
func Init() error {
	cfg := config.GetConfig()
	if !cfg.Focus.Enable {
		return nil
	}

	log = logger.GetLogger()

	var err error
	compiledPatterns, err = compilePatterns(&cfg.Focus)
	if err != nil {
		return fmt.Errorf("failed to compile regular expressions: %w", err)
	}

	return nil
}

// compilePatterns compiles regular expressions
func compilePatterns(focusConfig *config.FocusConfig) (*CompiledPatterns, error) {
	compiled := &CompiledPatterns{
		FilePatterns:   make([]*regexp.Regexp, 0),
		IgnorePatterns: make([]*regexp.Regexp, 0),
	}

	// Compile file path patterns
	for _, pattern := range focusConfig.FilePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile file pattern: %s, error: %v", pattern, err)
		}
		compiled.FilePatterns = append(compiled.FilePatterns, re)
	}

	// Compile ignore patterns
	for _, pattern := range focusConfig.IgnorePatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ignore pattern: %s, error: %v", pattern, err)
		}
		compiled.IgnorePatterns = append(compiled.IgnorePatterns, re)
	}

	return compiled, nil
}

// CheckFocusChange checks if a change should be marked as focus
func CheckFocusChange(change *types.ChangeInfo) (*types.FocusFileInfo, bool) {
	cfg := config.GetConfig()
	if !cfg.Focus.Enable {
		return nil, false
	}

	// Check if file is in ignore list (file path)
	if isIgnoredByFilePatterns(change.Filepath, compiledPatterns.IgnorePatterns) {
		return nil, false
	}

	// First step: check if file is yaml, yml, or json
	isTargetFile := false
	for _, pattern := range compiledPatterns.FilePatterns {
		if pattern.MatchString(change.Filepath) {
			isTargetFile = true
			break
		}
	}

	if !isTargetFile {
		return nil, false
	}

	focusFile := &types.FocusFileInfo{
		Filepath: change.Filepath,
		Action:   change.Action,
	}

	// Check for new files
	if cfg.Focus.AddFiles && change.Action == "add" {
		change.IsFocus = true
		change.FocusReason = "New file"
		focusFile.Reason = "New file"

		log.WithFields(logger.Fields{
			"file":   change.Filepath,
			"action": change.Action,
			"reason": change.FocusReason,
		}).Debug("New file marked as focus")

		return focusFile, true
	}

	// Check for deleted files
	if cfg.Focus.DeleteFiles && change.Action == "delete" {
		change.IsFocus = true
		change.FocusReason = "Deleted file"
		focusFile.Reason = "Deleted file"

		log.WithFields(logger.Fields{
			"file":   change.Filepath,
			"action": change.Action,
			"reason": change.FocusReason,
		}).Debug("Deleted file marked as focus")

		return focusFile, true
	}

	// Check for modified file content
	if cfg.Focus.ModifyFiles && change.Action == "modify" {
		matchedLines := make([]string, 0)
		matchCount := 0

		// Check added content: if a line doesn't match ignore patterns, mark as focus
		for _, line := range change.AdditionsList {
			// Check if this line doesn't contain any ignore patterns
			if !isLineIgnored(line.Content, compiledPatterns.IgnorePatterns) {
				matchCount++
				// Only save summary of matched line (first 100 characters)
				lineSummary := types.TruncateString(line.Content, 100)
				if len(lineSummary) > 0 && !types.Contains(matchedLines, lineSummary) {
					matchedLines = append(matchedLines, lineSummary)
				}
			}
		}

		// Check removed content: if a line doesn't match ignore patterns, mark as focus
		for _, line := range change.DeletionsList {
			// Check if this line doesn't contain any ignore patterns
			if !isLineIgnored(line.Content, compiledPatterns.IgnorePatterns) {
				matchCount++
				// Only save summary of matched line (first 100 characters)
				lineSummary := types.TruncateString(line.Content, 100)
				if len(lineSummary) > 0 && !types.Contains(matchedLines, lineSummary) {
					matchedLines = append(matchedLines, lineSummary)
				}
			}
		}

		if matchCount > 0 {
			change.IsFocus = true
			change.FocusReason = fmt.Sprintf("Content doesn't match ignore patterns, match count: %d", matchCount)
			focusFile.Reason = fmt.Sprintf("Content doesn't match ignore patterns, match count: %d", matchCount)
			focusFile.MatchCount = matchCount
			focusFile.MatchLines = matchedLines

			log.WithFields(logger.Fields{
				"file":        change.Filepath,
				"action":      change.Action,
				"match_count": matchCount,
				"reason":      change.FocusReason,
			}).Debug("Modified file content doesn't match ignore patterns")

			return focusFile, true
		}
	}

	return nil, false
}

// isIgnoredByFilePatterns checks if file matches ignore patterns
func isIgnoredByFilePatterns(filepath string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(filepath) {
			log.WithFields(logger.Fields{
				"file":    filepath,
				"pattern": pattern.String(),
			}).Debug("File matches ignore pattern")
			return true
		}
	}
	return false
}

// isLineIgnored checks if line content matches ignore patterns
func isLineIgnored(content string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(content) {
			return true
		}
	}
	return false
}
