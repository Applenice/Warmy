# Warmy
Warmy Git Commit Reader

A configuration-driven Git commit analysis tool with smart focus features.
Features  

✅ Smart commit analysis - Automatically analyzes Git commits and extracts complete change information  
🎯 Focus feature - Configurable intelligent focus system to identify critical file changes  
⚙️ Configuration-driven- All parameters managed through JSON configuration files  
📊 Rich output - Generates detailed JSON analysis reports  
🔍 Flexible filtering - Supports file type filtering and content pattern ignoring  
📁 Change type coverage - Independent focus configuration for new, modified, and deleted files  

### Installation

Compile from source

Go version: 1.24.11

#### Clone the repository
```shell
git clone https://github.com/Applenice/Warmy.git
cd Warmy
go build -o warmy
```

### Configuration File Explanation

#### Basic Settings

| Parameter | Value | Explanation |
|-----------|-------|-------------|
| **`repo_path`** | `"./"` | Specifies the repository path. The value `"./"` means the current directory. This tells the tool where to find the Git repository to analyze. |
| **`commit_hash`** | `""` | Specifies a particular commit hash to analyze. An empty string means the tool will analyze the latest commit. |
| **`output_dir`** | `"./analysis"` | The directory where analysis results will be saved. Results will be stored in a folder named "analysis" within the current directory. |
| **`pretty_json`** | `true` | Enables formatted, human-readable JSON output. If set to `false`, the JSON will be minified (more compact but less readable). |
| **`verbose`** | `false` | Controls whether verbose logging is enabled. When `true`, more detailed information is logged. Currently set to `false` for cleaner output. |
| **`parse_diff`** | `true` | Enables parsing of diff (difference) content. When `true`, the tool will analyze what specific lines were added, modified, or deleted in files. |
| **`no_file`** | `false` | Controls whether to prevent saving output to a file. When `false`, the tool will save results to the output directory. If `true`, results are only shown in console (if enabled). |
| **`no_console`** | `true` | Controls console output. When `true`, the tool will NOT display results in the console. Results will only be saved to file (since `no_file` is `false`). |
| **`log_level`** | `"info"` | Controls the verbosity of logs. `"info"` shows informational messages, warnings, and errors. Other options: `"debug"`, `"warn"`, `"error"`, `"fatal"`, `"panic"`. |
| **`max_diff_size`** | `1048576` | The maximum size (in bytes) of diff content to parse. This prevents memory issues with very large files. 1,048,576 bytes equals 1 MB. |

#### Focus Feature Settings

The focus feature allows you to intelligently identify important changes in specific types of files.

| Parameter | Value  | Explanation                                                                                                                                                                                                    |
|-----------|--------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`enable`** | `true` | Enables the focus feature. When `false`, the tool performs regular analysis without focus filtering.                                                                                                           |
| **`add_files`** | `true` | When enabled, newly added files that match the file patterns will be marked as "focus" (important).                                                                                                            |
| **`modify_files`** | `true` | When enabled, modified files that match the file patterns AND contain non-ignored changes will be marked as "focus".                                                                                           |
| **`delete_files`** | `true` | When enabled, deleted files that match the file patterns will be marked as "focus".                                                                                                                            |
| **`file_patterns`** | `[".*\\.yaml$", ".*\\.yml$", ".*\\.json$"]`   | Specify the file types that require attention, such as YAML.                                                                                                                                                   |
| **`ignore_patterns`** | `["digest"]`   | If a Git commit contains any of the listed keywords in its modified lines, it should be ignored. This is to filter out changes that do not require attention, such as those made by automated machine commits. |

### Usage
```shell
 ./warmy --config config.json
```
The generated analysis report is similar to: analysis/18d71446-20260108-001152.json

### Output Report Demo
```json
{
  "hash": "18d7144648d7474cac9dee02b104bc78ce47fc3d",
  "short_hash": "18d71446",
  "author": {
    "name": "Dhiyaneshwaran",
    "email": "leedhiyanesh@gmail.com",
    "when": "2026-01-07 17:42:53 +0530"
  },
  "committer": {
    "name": "GitHub",
    "email": "noreply@github.com",
    "when": "2026-01-07 17:42:53 +0530"
  },
  "message": "Change severity of CVE-2019-15823 to high",
  "description": "Updated severity level from critical to high for CVE-2019-15823.",
  "full_message": "Change severity of CVE-2019-15823 to high\n\nUpdated severity level from critical to high for CVE-2019-15823.",
  "parent_hashes": [
    "11d7a52653799333b5a2062393cef5a2f2b95b39"
  ],
  "changes": [
    {
      "action": "modify",
      "filepath": "http/cves/2019/CVE-2019-15823.yaml",
      "additions": 1,
      "deletions": 1,
      "diff_content": "diff --git a/http/cves/2019/CVE-2019-15823.yaml b/http/cves/2019/CVE-2019-15823.yaml\n--- a/http/cves/2019/CVE-2019-15823.yaml\n+++ b/http/cves/2019/CVE-2019-15823.yaml\n-  severity: critical\n+  severity: high\n",
      "extension": "yaml",
      "file_size": 1478,
      "additions_list": [
        {
          "type": "add",
          "content": "  severity: high"
        }
      ],
      "deletions_list": [
        {
          "type": "delete",
          "content": "  severity: critical"
        }
      ],
      "is_focus": true,
      "focus_reason": "Content doesn't match ignore patterns, match count: 2"
    }
  ],
  "focus_files": [
    {
      "filepath": "http/cves/2019/CVE-2019-15823.yaml",
      "action": "modify",
      "reason": "Content doesn't match ignore patterns, match count: 2",
      "match_count": 2,
      "match_lines": [
        "  severity: high",
        "  severity: critical"
      ]
    }
  ],
  "timestamp": 1767787973,
  "tree_hash": "ac7a748a48b495212a87d544d7a6132e4a9d058a",
  "files_changed": [
    "http/cves/2019/CVE-2019-15823.yaml"
  ],
  "stats": {
    "total_additions": 1,
    "total_deletions": 1,
    "total_files": 1,
    "add_files": 0,
    "delete_files": 0,
    "modify_files": 1,
    "rename_files": 0,
    "copy_files": 0,
    "binary_files": 0
  },
  "diff_summary": {
    "total_diff_size": 207,
    "max_diff_size": 1048576
  },
  "output_file": "18d71446-20260108-002302.json",
  "analyze_time": "20260108-002302",
  "focus_stats": {
    "total_focus_files": 1,
    "add_focus_files": 0,
    "modify_focus_files": 1,
    "delete_focus_files": 0,
    "match_pattern_files": 0,
    "match_content_files": 1
  }
}
```
