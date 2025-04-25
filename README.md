
# GitFame

**GitFame** is a command-line utility written in Go that analyzes developer contributions in Git repositories. It collects and aggregates commit author statistics such as the number of lines, unique commits, and the number of files changed.

This project was created as part of a training course, but it is a fully functional CLI tool that can be used in real-world scenarios.

---

## GitFame can:

- Analyze changes in a repository at a specific revision (`--revision`, defaults to `HEAD`).
- Count the number of lines, commits, and files for each author.
- Restrict the analysis by file extensions, languages, or glob patterns.
- Use the committer instead of the author for attribution (`--use-committer`).
- Output results in various formats: `tabular`, `csv`, `json`, `json-lines`.

Example output:
```
Name         Lines Commits Files
Alice Smith  1023  17      12
Bob Johnson  541   9       8
```

### Libraries used

- [`pflag`](https://github.com/spf13/pflag) — for enhanced CLI flag parsing.
- `os/exec` — to execute Git commands under the hood.
- `text/tabwriter`, `encoding/json`, `encoding/csv` — for flexible output formatting.

---

## Concurrency and Synchronization

To speed up processing of a large number of files, a **worker pool** is used with a configurable number of threads (`--cpu-count`).

Patterns and techniques used:
- **Worker Pool** — limits the number of goroutines running simultaneously.
- **WaitGroup** — waits for all tasks to complete.
- **Channels** — for error signaling and safe access to shared structures.
- **Manual locking with buffered channels** — used instead of standard mutexes to simplify and clarify the code.

---

## Example Usage

```bash
# Analyze all .go and .md files
gitfame --repository=. --extensions='.go,.md' --order-by=lines

# Analyze only Go and Markdown files
gitfame --languages='go,markdown' --format=json

# Use committer instead of author
gitfame --use-committer --format=csv

# Exclude vendor and testdata directories
gitfame --exclude='vendor/*,testdata/*'
```
