package main

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func CalculateStatisticsForFile(totalWorkers int, filesToCalculateStatistics []string,
	repository, revision string, useCommitter bool) map[string]programmerInfo {
	authorMap := make(map[string]programmerInfo)
	workersPool := make(chan struct{}, totalWorkers)
	wg := &sync.WaitGroup{}
	wg.Add(len(filesToCalculateStatistics))
	errCh := make(chan error, len(filesToCalculateStatistics))
	lock1 := make(chan struct{}, 1)
	lock2 := make(chan struct{}, 1)
	for _, file := range filesToCalculateStatistics {
		go func(file string) {
			defer func() {
				wg.Done()
				<-workersPool
			}()
			workersPool <- struct{}{}
			if err1 := calculateStatisticsForFile(file,
				repository, revision, authorMap, useCommitter, &lock1, &lock2); err1 != nil {
				errCh <- err1
			}
		}(file)
	}
	wg.Wait()
	close(errCh)
	if len(errCh) > 0 {
		for err := range errCh {
			_, err = os.Stderr.WriteString(err.Error())
			if err != nil {
				os.Exit(1)
			}
		}
		os.Exit(1)
	}
	return authorMap
}

func calculateStatisticsForFile(file, repository, revision string,
	authorMap map[string]programmerInfo, useCommitter bool,
	lock1, lock2 *chan struct{}) error {

	cmd := exec.Command("git", "blame", "--porcelain", revision, file)
	cmd.Dir = repository
	outBytes, err := cmd.Output()

	if len(outBytes) == 0 {
		err = runGitLog(revision, file, repository, authorMap, lock2)
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	lines := strings.Split(string(outBytes), "\n")
	commits := make(map[string]commitInfo)

	i := 0
	for i < len(lines) {
		currentCommitInfo, copyI := parseGitBlameOutput(lines, i, commits, useCommitter, lock1)
		if copyI == -1 {
			break
		}
		i = copyI

		if previousProgrammerInfo, ok := authorMap[currentCommitInfo.programmer]; !ok {
			curProgrammerInfo := programmerInfo{
				lines:   currentCommitInfo.linesInCommit,
				commits: make(map[string]struct{}),
				files:   make(map[string]struct{}),
			}
			curProgrammerInfo.commits[currentCommitInfo.commitHash] = struct{}{}
			curProgrammerInfo.files[file] = struct{}{}
			authorMap[currentCommitInfo.programmer] = curProgrammerInfo
		} else {
			previousProgrammerInfo.lines += currentCommitInfo.linesInCommit
			previousProgrammerInfo.commits[currentCommitInfo.commitHash] = struct{}{}
			previousProgrammerInfo.files[file] = struct{}{}

			authorMap[currentCommitInfo.programmer] = previousProgrammerInfo
		}
		<-*lock1
	}
	return nil
}

func parseGitBlameOutput(lines []string, i int,
	commits map[string]commitInfo, useCommitter bool, lock *chan struct{}) (commitInfo, int) {
	firstLine := strings.Fields(lines[i])
	if len(firstLine) <= 1 {
		return commitInfo{}, -1
	}
	commitHash := firstLine[0]
	if len(firstLine) < 4 {
		return commitInfo{}, -1
	}
	linesInCommit, _ := strconv.Atoi(firstLine[3])
	var nextCommitLineNumber int
	_, ok := commits[commitHash]
	if ok {
		nextCommitLineNumber = i + linesInCommit*2
	} else {
		nextCommitLineNumber = i + 10
		for {
			if strings.Fields(lines[nextCommitLineNumber])[0] == "filename" {
				break
			}
			nextCommitLineNumber++
		}
		nextCommitLineNumber += linesInCommit * 2
	}

	var programmerName string
	if ok {
		programmerName = commits[commitHash].programmer
	} else {
		i += 1
		author := lines[i][len("author "):]
		i += 4
		commiter := lines[i][len("committer "):]
		if !useCommitter {
			programmerName = author
		} else {
			programmerName = commiter
		}
	}
	i = nextCommitLineNumber

	answer := commitInfo{
		commitHash:    commitHash,
		programmer:    programmerName,
		linesInCommit: linesInCommit,
	}
	*lock <- struct{}{}
	if _, ok = commits[commitHash]; !ok {
		commits[commitHash] = answer
	}
	return answer, i
}

type commitInfo struct {
	commitHash    string
	programmer    string
	linesInCommit int
}

type programmerInfo struct {
	lines   int
	commits map[string]struct{}
	files   map[string]struct{}
}

func runGitLog(revision, file, repository string,
	authorMap map[string]programmerInfo, lock *chan struct{}) error {
	cmd2 := exec.Command("git", "log", revision, "--", file)

	cmd2.Dir = repository
	outBytes, err := cmd2.Output()

	if err != nil {
		return err
	}

	lines := strings.Split(string(outBytes), "\n")

	commitHash := strings.Fields(lines[0])[1]
	var commitAuthor string
	authorStringFields := strings.Fields(lines[1])
	authorMail := authorStringFields[len(authorStringFields)-1]

	for i := range lines[1] {
		if i+len(authorMail) == len(lines[1])+1 {
			break
		}

		if lines[1][i:i+len(authorMail)] == authorMail {
			commitAuthor = lines[1][len("Author: ") : i-1]
			for j := range commitAuthor {
				if commitAuthor[j] == '"' {
					commitAuthor = "\"" + commitAuthor[:j] + "\"" + commitAuthor[j:] + "\""
					break
				}
			}
		}
	}
	*lock <- struct{}{}
	previousAuthorInfo, ok := authorMap[commitAuthor]
	if ok {
		previousAuthorInfo.commits[commitHash] = struct{}{}
		previousAuthorInfo.files[file] = struct{}{}
		authorMap[commitAuthor] = previousAuthorInfo
	} else {
		info := programmerInfo{
			lines:   0,
			commits: make(map[string]struct{}),
			files:   make(map[string]struct{}),
		}
		info.files[file] = struct{}{}
		info.commits[commitHash] = struct{}{}
		authorMap[commitAuthor] = info
	}
	<-*lock
	return nil
}
