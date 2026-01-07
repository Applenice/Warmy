package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"

	"warmy/internal/config"
	"warmy/internal/focus"
	"warmy/internal/logger"
	"warmy/internal/types"
)

var log logger.Logger

// GetCommit gets complete information of specified commit
func GetCommit(repoPath, commitHash string) (*types.CommitInfo, error) {
	log = logger.GetLogger()

	log.WithFields(logger.Fields{
		"repo_path":   repoPath,
		"commit_hash": commitHash,
	}).Info("Started reading specified commit")

	// Open local repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.WithFields(logger.Fields{
			"repo_path": repoPath,
			"error":     err.Error(),
		}).Error("Failed to open local repository")
		return nil, fmt.Errorf("failed to open local repository: %w", err)
	}

	log.Debug("Successfully opened local repository")

	var commit *object.Commit

	if commitHash == "" {
		// If no commit hash specified, get latest commit
		ref, err := repo.Head()
		if err != nil {
			log.WithFields(logger.Fields{
				"repo_path": repoPath,
				"error":     err.Error(),
			}).Error("Failed to get HEAD reference")
			return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
		}

		log.WithFields(logger.Fields{
			"ref":  ref.Name().String(),
			"hash": ref.Hash().String(),
		}).Debug("Got HEAD reference")

		// Get commit object
		commit, err = repo.CommitObject(ref.Hash())
		if err != nil {
			log.WithFields(logger.Fields{
				"hash":  ref.Hash().String(),
				"error": err.Error(),
			}).Error("Failed to get commit object")
			return nil, fmt.Errorf("failed to get commit object: %w", err)
		}
	} else {
		// Parse specified commit hash
		hash := plumbing.NewHash(commitHash)
		commit, err = repo.CommitObject(hash)
		if err != nil {
			// Try to find short hash
			commitIter, err := repo.CommitObjects()
			if err != nil {
				log.WithFields(logger.Fields{
					"hash":  commitHash,
					"error": err.Error(),
				}).Error("Failed to iterate commit objects")
				return nil, fmt.Errorf("failed to iterate commit objects: %w", err)
			}

			var foundCommit *object.Commit
			err = commitIter.ForEach(func(c *object.Commit) error {
				if strings.HasPrefix(c.Hash.String(), commitHash) {
					foundCommit = c
					return fmt.Errorf("found") // Break iteration
				}
				return nil
			})

			if foundCommit != nil {
				commit = foundCommit
			} else if err != nil && err.Error() != "found" {
				log.WithFields(logger.Fields{
					"hash":  commitHash,
					"error": err.Error(),
				}).Error("Failed to find commit")
				return nil, fmt.Errorf("failed to find commit: %w", err)
			} else {
				log.WithFields(logger.Fields{
					"hash": commitHash,
				}).Error("Specified commit not found")
				return nil, fmt.Errorf("specified commit not found: %s", commitHash)
			}
		}
	}

	log.WithFields(logger.Fields{
		"commit_hash": commit.Hash.String(),
		"author":      commit.Author.Name,
		"message":     strings.Split(commit.Message, "\n")[0],
	}).Info("Got commit object")

	// Get branch information
	branches, err := getBranchesContainingCommit(repo, commit.Hash)
	if err != nil {
		log.WithError(err).Warn("Failed to get branch information")
		branches = []string{}
	}

	// Get tag information
	tags, err := getTagsContainingCommit(repo, commit.Hash)
	if err != nil {
		log.WithError(err).Warn("Failed to get tag information")
		tags = []string{}
	}

	// Parse commit message
	message := strings.TrimSpace(commit.Message)
	subject, description := types.SplitCommitMessage(message)

	log.WithFields(logger.Fields{
		"subject_length":     len(subject),
		"description_length": len(description),
	}).Debug("Parsed commit message")

	// Get parent commit hashes
	parentHashes := make([]string, 0)
	if commit.NumParents() > 0 {
		err = commit.Parents().ForEach(func(p *object.Commit) error {
			parentHashes = append(parentHashes, p.Hash.String())
			return nil
		})
		if err != nil {
			log.WithFields(logger.Fields{
				"commit": commit.Hash.String(),
				"error":  err.Error(),
			}).Warn("Failed to iterate parent commits, continuing")
		}
	}

	log.WithFields(logger.Fields{
		"parent_count":      len(parentHashes),
		"is_initial_commit": commit.NumParents() == 0,
	}).Debug("Got parent commit information")

	// Get commit tree object
	tree, err := commit.Tree()
	if err != nil {
		log.WithFields(logger.Fields{
			"commit": commit.Hash.String(),
			"error":  err.Error(),
		}).Error("Failed to get tree object")
		return nil, fmt.Errorf("failed to get tree object: %w", err)
	}

	log.WithFields(logger.Fields{
		"tree_hash": tree.Hash.String(),
	}).Debug("Got tree object")

	// Get change information
	changes, stats, diffSummary, err := getCommitChanges(repo, commit)
	if err != nil {
		log.WithFields(logger.Fields{
			"commit": commit.Hash.String(),
			"error":  err.Error(),
		}).Warn("Failed to get change information, returning empty change list")
		changes = []types.ChangeInfo{}
		stats = types.StatsInfo{}
		diffSummary = types.DiffSummary{}
	} else {
		log.WithFields(logger.Fields{
			"total_files":     stats.TotalFiles,
			"additions":       stats.TotalAdditions,
			"deletions":       stats.TotalDeletions,
			"binary_files":    stats.BinaryFiles,
			"total_diff_size": diffSummary.TotalDiffSize,
		}).Info("Successfully got change information")
	}

	// Build changed file list
	filesChanged := make([]string, 0, len(changes))
	focusFiles := make([]types.FocusFileInfo, 0)
	focusStats := types.FocusStats{}

	// Initialize focus feature
	if err := focus.Init(); err != nil {
		log.WithError(err).Warn("Failed to initialize focus feature")
	} else {
		for i := range changes {
			change := &changes[i]
			filesChanged = append(filesChanged, change.Filepath)

			// Check if change is focus
			if focusFile, isFocus := focus.CheckFocusChange(change); isFocus {
				focusFiles = append(focusFiles, *focusFile)

				// Statistics
				focusStats.TotalFocusFiles++
				if change.Action == "add" {
					focusStats.AddFocusFiles++
					focusStats.MatchPatternFiles++
				} else if change.Action == "modify" {
					focusStats.ModifyFocusFiles++
					focusStats.MatchContentFiles++
				} else if change.Action == "delete" {
					// Delete files don't have content to match, so they're counted as pattern files
					focusStats.MatchPatternFiles++
				}
			}
		}
	}

	log.WithFields(logger.Fields{
		"file_count":  len(filesChanged),
		"focus_count": len(focusFiles),
	}).Debug("Built changed file list")

	// Get current time
	currentTime := time.Now()
	analyzeTime := currentTime.Format("20060102-150405")

	// Build commit information
	commitInfo := &types.CommitInfo{
		Hash:      commit.Hash.String(),
		ShortHash: commit.Hash.String()[:8], // Take first 8 characters as short hash
		Author: types.AuthorInfo{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			When:  commit.Author.When.Format("2006-01-02 15:04:05 -0700"),
		},
		Committer: types.AuthorInfo{
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
			When:  commit.Committer.When.Format("2006-01-02 15:04:05 -0700"),
		},
		Message:      subject,
		Description:  description,
		FullMessage:  message,
		ParentHashes: parentHashes,
		Changes:      changes,
		FocusFiles:   focusFiles,
		Timestamp:    commit.Committer.When.Unix(),
		TreeHash:     tree.Hash.String(),
		FilesChanged: filesChanged,
		Stats:        stats,
		DiffSummary:  diffSummary,
		Branches:     branches,
		Tags:         tags,
		AnalyzeTime:  analyzeTime,
		FocusStats:   focusStats,
	}

	log.WithFields(logger.Fields{
		"commit_hash":  commitInfo.Hash,
		"message":      commitInfo.Message,
		"author":       commitInfo.Author.Name,
		"total_files":  commitInfo.Stats.TotalFiles,
		"focus_files":  len(commitInfo.FocusFiles),
		"branches":     len(commitInfo.Branches),
		"tags":         len(commitInfo.Tags),
		"analyze_time": commitInfo.AnalyzeTime,
	}).Info("Successfully built commit information")

	return commitInfo, nil
}

// getBranchesContainingCommit gets branches containing specified commit
func getBranchesContainingCommit(repo *git.Repository, hash plumbing.Hash) ([]string, error) {
	branches := []string{}

	// Get all branch references
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Only process local and remote branches
		if ref.Name().IsBranch() || strings.HasPrefix(ref.Name().String(), "refs/remotes/") {
			// Simplified: only check if current reference points to this commit
			if ref.Hash() == hash {
				branches = append(branches, ref.Name().Short())
			}
		}
		return nil
	})

	return branches, err
}

// getTagsContainingCommit gets tags containing specified commit
func getTagsContainingCommit(repo *git.Repository, hash plumbing.Hash) ([]string, error) {
	tags := []string{}

	// Get all tag references
	refs, err := repo.References()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		// Only process tags
		if ref.Name().IsTag() {
			// If annotated tag, need to parse tag object
			var tagCommit plumbing.Hash

			obj, err := repo.TagObject(ref.Hash())
			if err == nil {
				// Annotated tag
				tagCommit = obj.Target
			} else if err == plumbing.ErrObjectNotFound {
				// Lightweight tag
				tagCommit = ref.Hash()
			} else {
				return err
			}

			if tagCommit == hash {
				tags = append(tags, ref.Name().Short())
			}
		}
		return nil
	})

	return tags, err
}

// getCommitChanges gets change information of commit
func getCommitChanges(repo *git.Repository, commit *object.Commit) ([]types.ChangeInfo, types.StatsInfo, types.DiffSummary, error) {
	changes := make([]types.ChangeInfo, 0)
	stats := types.StatsInfo{}
	cfg := config.GetConfig()
	diffSummary := types.DiffSummary{
		MaxDiffSize: cfg.MaxDiffSize,
	}

	log := logger.GetLogger().WithFields(logger.Fields{
		"commit": commit.Hash.String(),
	})

	log.Debug("Started getting change information")

	// If initial commit, no parent
	if commit.NumParents() == 0 {
		log.Info("This is initial commit, getting all files")

		// Get all files in tree
		tree, err := commit.Tree()
		if err != nil {
			log.WithError(err).Error("Failed to get tree object")
			return changes, stats, diffSummary, err
		}

		// Iterate all files
		fileCount := 0
		totalDiffSize := 0
		err = tree.Files().ForEach(func(f *object.File) error {
			// Get file size
			size := f.Size

			// Get file content
			content, err := f.Contents()
			var diffContent string
			if err != nil {
				diffContent = fmt.Sprintf("// Unable to read file content: %v\n", err)
				stats.BinaryFiles++
			} else {
				// Generate diff for initial commit (full file content)
				diffContent = fmt.Sprintf("+++ b/%s\n@@ -0,0 +1,%d @@\n%s",
					f.Name, len(strings.Split(content, "\n")), content)
			}

			change := types.ChangeInfo{
				Action:      "add",
				Filepath:    f.Name,
				Additions:   len(strings.Split(content, "\n")),
				Deletions:   0,
				DiffContent: diffContent,
				Extension:   types.GetFileExtension(f.Name),
				FileSize:    size,
				IsBinary:    err != nil, // If cannot read content, might be binary file
			}

			// Parse diff content
			if cfg.ParseDiff && !change.IsBinary {
				additions, _ := parseDiffContent(diffContent)
				change.AdditionsList = additions
			}

			changes = append(changes, change)
			stats.AddFiles++
			fileCount++

			// Count diff size
			diffSize := len(diffContent)
			totalDiffSize += diffSize

			// Check if single diff is too large
			if diffSize > cfg.MaxDiffSize {
				change.DiffContent = fmt.Sprintf("// Diff content too large (%d bytes), truncated", diffSize)
				diffSummary.DiffTooLarge = true
			}

			return nil
		})

		if err != nil {
			log.WithError(err).Error("Failed to iterate files")
			return changes, stats, diffSummary, err
		}

		stats.TotalFiles = len(changes)
		diffSummary.TotalDiffSize = totalDiffSize

		log.WithFields(logger.Fields{
			"total_files":     fileCount,
			"total_diff_size": totalDiffSize,
		}).Debug("Initial commit file statistics completed")

		return changes, stats, diffSummary, nil
	}

	// Get first parent commit (usually the most direct previous commit)
	parent, err := commit.Parent(0)
	if err != nil {
		log.WithError(err).Error("Failed to get parent commit")
		return changes, stats, diffSummary, err
	}

	log.WithFields(logger.Fields{
		"parent_hash": parent.Hash.String(),
	}).Debug("Got parent commit")

	// Get parent commit tree
	parentTree, err := parent.Tree()
	if err != nil {
		log.WithError(err).Error("Failed to get parent commit tree")
		return changes, stats, diffSummary, err
	}

	// Get current commit tree
	currentTree, err := commit.Tree()
	if err != nil {
		log.WithError(err).Error("Failed to get current commit tree")
		return changes, stats, diffSummary, err
	}

	log.Debug("Started generating patch")

	// Compare two trees
	patch, err := parentTree.Patch(currentTree)
	if err != nil {
		log.WithError(err).Error("Failed to generate patch")
		return changes, stats, diffSummary, err
	}

	log.WithFields(logger.Fields{
		"patch_files": len(patch.FilePatches()),
	}).Debug("Patch generation completed")

	// Process each file change
	fullDiff := ""
	totalDiffSize := 0

	for i, filePatch := range patch.FilePatches() {
		fromFile, toFile := filePatch.Files()

		change := types.ChangeInfo{}
		var filePath string
		var fromPath, toPath string

		// Get file path
		if fromFile != nil {
			fromPath = fromFile.Path()
		}
		if toFile != nil {
			toPath = toFile.Path()
		}

		// Determine change type and file path
		if fromFile == nil && toFile != nil {
			// Added file
			change.Action = "add"
			filePath = toPath
			change.Filepath = filePath
			stats.AddFiles++
		} else if fromFile != nil && toFile == nil {
			// Deleted file
			change.Action = "delete"
			filePath = fromPath
			change.Filepath = filePath
			stats.DeleteFiles++
		} else if fromFile != nil && toFile != nil {
			// Modified, renamed or copied
			if fromPath != toPath {
				// Renamed
				change.Action = "rename"
				change.OldPath = fromPath
				change.NewPath = toPath
				change.Filepath = toPath
				filePath = toPath
				stats.RenameFiles++

				log.WithFields(logger.Fields{
					"file_index": i,
					"old_path":   fromPath,
					"new_path":   toPath,
				}).Debug("Detected file rename")
			} else {
				// Modified
				change.Action = "modify"
				filePath = fromPath
				change.Filepath = filePath
				stats.ModifyFiles++
			}
		}

		// Get file extension
		change.Extension = types.GetFileExtension(filePath)

		// Count line changes and generate diff content
		additions := 0
		deletions := 0

		// Generate diff content
		var diffContentBuilder strings.Builder

		// Write diff header
		if fromFile == nil && toFile != nil {
			// Added file
			diffContentBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", toPath, toPath))
			diffContentBuilder.WriteString(fmt.Sprintf("new file mode 100644\n"))
			diffContentBuilder.WriteString(fmt.Sprintf("--- /dev/null\n"))
			diffContentBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", toPath))
		} else if fromFile != nil && toFile == nil {
			// Deleted file
			diffContentBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fromPath, fromPath))
			diffContentBuilder.WriteString(fmt.Sprintf("deleted file mode 100644\n"))
			diffContentBuilder.WriteString(fmt.Sprintf("--- a/%s\n", fromPath))
			diffContentBuilder.WriteString(fmt.Sprintf("+++ /dev/null\n"))
		} else if fromFile != nil && toFile != nil {
			if fromPath == toPath {
				// Modified file
				diffContentBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fromPath, toPath))
				diffContentBuilder.WriteString(fmt.Sprintf("--- a/%s\n", fromPath))
				diffContentBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", toPath))
			} else {
				// Renamed file
				diffContentBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", fromPath, toPath))
				diffContentBuilder.WriteString(fmt.Sprintf("rename from %s\n", fromPath))
				diffContentBuilder.WriteString(fmt.Sprintf("rename to %s\n", toPath))
			}
		}

		// Process each chunk
		for _, chunk := range filePatch.Chunks() {
			content := chunk.Content()
			lines := strings.Split(content, "\n")

			// Remove trailing empty string (if exists)
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				lines = lines[:len(lines)-1]
			}

			lineCount := len(lines)

			switch chunk.Type() {
			case diff.Add:
				// Added lines
				additions += lineCount
				stats.TotalAdditions += lineCount

				// Add added lines to diff
				for _, line := range lines {
					if line != "" {
						diffContentBuilder.WriteString(fmt.Sprintf("+%s\n", line))
					}
				}

			case diff.Delete:
				// Deleted lines
				deletions += lineCount
				stats.TotalDeletions += lineCount

				// Add deleted lines to diff
				for _, line := range lines {
					if line != "" {
						diffContentBuilder.WriteString(fmt.Sprintf("-%s\n", line))
					}
				}
			}
		}

		change.Additions = additions
		change.Deletions = deletions

		// Get detailed diff content
		if toFile != nil {
			// Try to get file size
			file, err := currentTree.File(filePath)
			if err == nil {
				change.FileSize = file.Size
			}

			// Check if file is likely binary
			change.IsBinary = types.IsLikelyBinaryFile(filePath)

			// Get generated diff content
			fileDiff := diffContentBuilder.String()
			change.DiffContent = fileDiff

			if change.IsBinary {
				stats.BinaryFiles++
			} else {
				// Parse diff content
				if cfg.ParseDiff {
					additions, deletions := parseDiffContent(fileDiff)
					change.AdditionsList = additions
					change.DeletionsList = deletions
				}
			}

			// Check diff size
			diffSize := len(fileDiff)
			totalDiffSize += diffSize

			if diffSize > cfg.MaxDiffSize {
				change.DiffContent = fmt.Sprintf("// Diff content too large (%d bytes), truncated", diffSize)
				diffSummary.DiffTooLarge = true
			}

			// If full file diff is needed, add to fullDiff
			if cfg.IncludeFullDiff {
				fullDiff += fileDiff + "\n\n"
			}
		} else if fromFile != nil {
			// Deleted file case
			change.IsBinary = types.IsLikelyBinaryFile(filePath)
			if change.IsBinary {
				stats.BinaryFiles++
			} else {
				// Parse diff content
				if cfg.ParseDiff {
					fileDiff := diffContentBuilder.String()
					_, deletions := parseDiffContent(fileDiff)
					change.DeletionsList = deletions
				}
			}

			// Get generated diff content
			fileDiff := diffContentBuilder.String()
			change.DiffContent = fileDiff

			diffSize := len(fileDiff)
			totalDiffSize += diffSize

			if cfg.IncludeFullDiff {
				fullDiff += fileDiff + "\n\n"
			}
		}

		changes = append(changes, change)

		// Log detailed change information
		if log.GetLevel() >= logrus.DebugLevel {
			log.WithFields(logger.Fields{
				"file_index":      i,
				"file":            filePath,
				"action":          change.Action,
				"additions":       change.Additions,
				"deletions":       change.Deletions,
				"diff_size":       len(change.DiffContent),
				"is_binary":       change.IsBinary,
				"additions_count": len(change.AdditionsList),
				"deletions_count": len(change.DeletionsList),
			}).Debug("File change details")
		}
	}

	stats.TotalFiles = len(changes)
	diffSummary.TotalDiffSize = totalDiffSize

	if cfg.IncludeFullDiff {
		diffSummary.FullDiff = fullDiff
	}

	log.WithFields(logger.Fields{
		"total_files":     stats.TotalFiles,
		"add_files":       stats.AddFiles,
		"delete_files":    stats.DeleteFiles,
		"modify_files":    stats.ModifyFiles,
		"rename_files":    stats.RenameFiles,
		"binary_files":    stats.BinaryFiles,
		"total_additions": stats.TotalAdditions,
		"total_deletions": stats.TotalDeletions,
		"total_diff_size": diffSummary.TotalDiffSize,
		"diff_too_large":  diffSummary.DiffTooLarge,
	}).Info("Change information statistics completed")

	return changes, stats, diffSummary, nil
}

// parseDiffContent parses diff string, extracts added and deleted lines
func parseDiffContent(diffContent string) ([]types.LineChange, []types.LineChange) {
	var additions []types.LineChange
	var deletions []types.LineChange

	if diffContent == "" {
		return additions, deletions
	}

	// Split diff by lines
	lines := strings.Split(diffContent, "\n")
	if len(lines) < 3 { // diff should have at least 3 lines
		return additions, deletions
	}

	// Check if has standard hunk header (@@ lines)
	hasHunkHeader := false
	inHunk := false

	// First check if has hunk
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			hasHunkHeader = true
			break
		}
	}

	if hasHunkHeader {
		// Has standard hunk header, parse in standard way
		for _, line := range lines {
			// Skip diff header
			if strings.HasPrefix(line, "diff --git") ||
				strings.HasPrefix(line, "index ") ||
				strings.HasPrefix(line, "--- ") ||
				strings.HasPrefix(line, "+++ ") ||
				strings.HasPrefix(line, "new file mode") ||
				strings.HasPrefix(line, "deleted file mode") ||
				strings.HasPrefix(line, "rename from") ||
				strings.HasPrefix(line, "rename to") {
				continue
			}

			// Check if is hunk header
			if strings.HasPrefix(line, "@@") {
				inHunk = true
				// Skip hunk header, no need to parse
				continue
			} else if inHunk && len(line) > 0 {
				// Process line
				prefix := line[0:1]
				content := line[1:]

				switch prefix {
				case "+": // Added line
					// Skip diff's +++ line (file header)
					if !strings.HasPrefix(content, "++ b/") {
						additions = append(additions, types.LineChange{
							Type:    "add",
							Content: content,
						})
					}

				case "-": // Deleted line
					// Skip diff's --- line (file header)
					if !strings.HasPrefix(content, "-- a/") {
						deletions = append(deletions, types.LineChange{
							Type:    "delete",
							Content: content,
						})
					}

				case " ": // Context line, skip
					continue
				}
			}
		}
	} else {
		// No standard hunk header, use simple parsing
		// Only parse actual change lines, skip file headers
		skipHeader := false
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}

			// Skip diff header
			if strings.HasPrefix(line, "diff --git") ||
				strings.HasPrefix(line, "index ") ||
				strings.HasPrefix(line, "--- ") ||
				strings.HasPrefix(line, "+++ ") ||
				strings.HasPrefix(line, "new file mode") ||
				strings.HasPrefix(line, "deleted file mode") ||
				strings.HasPrefix(line, "rename from") ||
				strings.HasPrefix(line, "rename to") {
				skipHeader = true
				continue
			}

			// If encountered non-header line, start parsing
			if skipHeader && len(line) > 0 {
				prefix := line[0:1]
				content := line[1:]

				switch prefix {
				case "+": // Added line
					// Skip diff's +++ line (file header)
					if !strings.HasPrefix(content, "++ b/") {
						additions = append(additions, types.LineChange{
							Type:    "add",
							Content: content,
						})
					}

				case "-": // Deleted line
					// Skip diff's --- line (file header)
					if !strings.HasPrefix(content, "-- a/") {
						deletions = append(deletions, types.LineChange{
							Type:    "delete",
							Content: content,
						})
					}
				}
			}
		}
	}

	// Log parsing result
	log.WithFields(logger.Fields{
		"additions_count": len(additions),
		"deletions_count": len(deletions),
	}).Debug("Diff content parsing completed")

	return additions, deletions
}
