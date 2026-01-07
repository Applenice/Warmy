package types

import (
	"encoding/json"
	"strings"
)

// LineChange represents a specific changed line
type LineChange struct {
	Type    string `json:"type"`    // Change type: "add" or "delete"
	Content string `json:"content"` // Line content
}

// ChangeInfo represents file change information
type ChangeInfo struct {
	Action        string       `json:"action"`                   // Change type: add, delete, modify, rename, copy
	Filepath      string       `json:"filepath"`                 // File path
	OldPath       string       `json:"old_path,omitempty"`       // Original path for rename/copy
	NewPath       string       `json:"new_path,omitempty"`       // New path for rename/copy
	Additions     int          `json:"additions"`                // Number of added lines
	Deletions     int          `json:"deletions"`                // Number of deleted lines
	DiffContent   string       `json:"diff_content,omitempty"`   // Original diff content
	Extension     string       `json:"extension,omitempty"`      // File extension
	FileSize      int64        `json:"file_size,omitempty"`      // File size (bytes)
	IsBinary      bool         `json:"is_binary,omitempty"`      // Whether it's a binary file
	AdditionsList []LineChange `json:"additions_list,omitempty"` // Added lines
	DeletionsList []LineChange `json:"deletions_list,omitempty"` // Deleted lines
	IsFocus       bool         `json:"is_focus,omitempty"`       // Whether it's a focus file
	FocusReason   string       `json:"focus_reason,omitempty"`   // Focus reason
}

// FocusFileInfo represents focus file information
type FocusFileInfo struct {
	Filepath   string   `json:"filepath"`              // File path
	Action     string   `json:"action"`                // Change type
	Reason     string   `json:"reason"`                // Focus reason
	MatchCount int      `json:"match_count,omitempty"` // Number of matched content
	MatchLines []string `json:"match_lines,omitempty"` // Matched line content (summary)
}

// AuthorInfo represents author/committer information
type AuthorInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	When  string `json:"when"`
}

// StatsInfo represents change statistics
type StatsInfo struct {
	TotalAdditions int `json:"total_additions"` // Total added lines
	TotalDeletions int `json:"total_deletions"` // Total deleted lines
	TotalFiles     int `json:"total_files"`     // Total changed files
	AddFiles       int `json:"add_files"`       // Number of added files
	DeleteFiles    int `json:"delete_files"`    // Number of deleted files
	ModifyFiles    int `json:"modify_files"`    // Number of modified files
	RenameFiles    int `json:"rename_files"`    // Number of renamed files
	CopyFiles      int `json:"copy_files"`      // Number of copied files
	BinaryFiles    int `json:"binary_files"`    // Number of binary files
}

// FocusStats represents focus statistics
type FocusStats struct {
	TotalFocusFiles   int `json:"total_focus_files"`   // Total focus files
	AddFocusFiles     int `json:"add_focus_files"`     // Number of new focus files
	ModifyFocusFiles  int `json:"modify_focus_files"`  // Number of modified focus files
	DeleteFocusFiles  int `json:"delete_focus_files"`  // Number of deleted focus files
	MatchPatternFiles int `json:"match_pattern_files"` // Number of files matching pattern
	MatchContentFiles int `json:"match_content_files"` // Number of files matching content
}

// DiffSummary represents diff summary information
type DiffSummary struct {
	TotalDiffSize int    `json:"total_diff_size"`          // Total diff size
	DiffTooLarge  bool   `json:"diff_too_large,omitempty"` // Whether diff is too large
	MaxDiffSize   int    `json:"max_diff_size,omitempty"`  // Maximum diff size limit
	FullDiff      string `json:"full_diff,omitempty"`      // Complete diff content
}

// CommitInfo represents complete commit information
type CommitInfo struct {
	Hash         string          `json:"hash"`                   // Commit hash
	ShortHash    string          `json:"short_hash"`             // Short hash
	Author       AuthorInfo      `json:"author"`                 // Author information
	Committer    AuthorInfo      `json:"committer"`              // Committer information
	Message      string          `json:"message"`                // Commit message subject
	Description  string          `json:"description"`            // Detailed description
	FullMessage  string          `json:"full_message"`           // Full commit message
	ParentHashes []string        `json:"parent_hashes"`          // Parent commit hash list
	Changes      []ChangeInfo    `json:"changes"`                // Change content list
	FocusFiles   []FocusFileInfo `json:"focus_files,omitempty"`  // Focus change file list
	Timestamp    int64           `json:"timestamp"`              // Commit timestamp
	TreeHash     string          `json:"tree_hash"`              // Tree object hash
	FilesChanged []string        `json:"files_changed"`          // Changed file list
	Stats        StatsInfo       `json:"stats"`                  // Statistics
	DiffSummary  DiffSummary     `json:"diff_summary"`           // Diff summary
	Branches     []string        `json:"branches,omitempty"`     // Belonging branches
	Tags         []string        `json:"tags,omitempty"`         // Tags
	OutputFile   string          `json:"output_file,omitempty"`  // Output file path
	AnalyzeTime  string          `json:"analyze_time,omitempty"` // Analysis time
	FocusStats   FocusStats      `json:"focus_stats,omitempty"`  // Focus statistics
}

// ToJSON converts CommitInfo to JSON string
func (c *CommitInfo) ToJSON(pretty bool) (string, error) {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(c, "", "  ")
	} else {
		data, err = json.Marshal(c)
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SplitCommitMessage splits commit message into subject and description
func SplitCommitMessage(message string) (string, string) {
	// Split first line as subject, rest as description
	lines := strings.Split(message, "\n")
	if len(lines) == 0 {
		return "", ""
	}

	subject := strings.TrimSpace(lines[0])
	description := ""

	if len(lines) > 1 {
		// Skip first empty line (if exists)
		start := 1
		if start < len(lines) && lines[1] == "" {
			start = 2
		}

		descLines := make([]string, 0)
		for i := start; i < len(lines); i++ {
			descLines = append(descLines, strings.TrimSpace(lines[i]))
		}
		description = strings.Join(descLines, "\n")
	}

	return subject, description
}

// GetFileExtension gets file extension
func GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// IsLikelyBinaryFile checks if file is likely binary
func IsLikelyBinaryFile(filename string) bool {
	// Common binary file extensions
	binaryExtensions := map[string]bool{
		"exe": true, "dll": true, "so": true, "dylib": true,
		"bin": true, "class": true, "jar": true, "war": true,
		"png": true, "jpg": true, "jpeg": true, "gif": true,
		"bmp": true, "ico": true, "pdf": true, "doc": true,
		"docx": true, "xls": true, "xlsx": true, "ppt": true,
		"pptx": true, "zip": true, "tar": true, "gz": true,
		"7z": true, "rar": true, "mp3": true, "mp4": true,
		"avi": true, "mkv": true, "mov": true, "wav": true,
		"iso": true, "img": true, "dmg": true, "pkg": true,
		"o": true, "obj": true, "lib": true, "a": true,
	}

	ext := GetFileExtension(filename)
	ext = strings.ToLower(ext)
	return binaryExtensions[ext]
}

// TruncateString truncates string
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... [content truncated]"
}

// Contains checks if string slice contains an item
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
