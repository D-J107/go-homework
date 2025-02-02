package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

func findMatchFiles(repository, revision string, extensions,
	languages, exclude, restrictTo []string) ([]string, error) {

	cmd := exec.Command("git", "ls-tree", "-r", revision)
	cmd.Dir = repository
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	_ = cmd.Run()

	out := outBuffer.String()

	files := make([]string, 0)
	for _, line := range strings.Split(out, "\n") {
		lineContents := strings.Fields(line)
		if len(lineContents) < 4 {
			continue
		}
		if lineContents[1] == "blob" {
			fileName := lineContents[3]
			for i := range line {
				if line[i:i+len(fileName)] == fileName {
					files = append(files, line[i:])
					break
				}
			}
		}
	}

	files = applyFilter(files, extensions)

	languagesExtensions, err := getLanguagesExtension(languages)
	if err != nil {
		return nil, err
	}
	files = applyFilter(files, languagesExtensions)
	files = applyGlobFilter(files, exclude, true)
	files = applyGlobFilter(files, restrictTo, false)
	return files, nil
}

func applyFilter(files, regexs []string) []string {
	if len(regexs) == 0 {
		return files
	}
	filteredFiles := make([]string, 0)
	for _, file := range files {
		for _, regex := range regexs {
			if strings.HasSuffix(file, regex) {
				filteredFiles = append(filteredFiles, file)
				break
			}
		}
	}
	return filteredFiles
}

func applyGlobFilter(files, globs []string, exclude bool) []string {
	if len(globs) == 0 {
		return files
	}

	filteredFiles := make([]string, 0)
	for _, file := range files {
		shouldAppend := exclude

		for _, pattern := range globs {

			matched, err := filepath.Match(pattern, file)
			if err != nil {
				log.Fatal(err)
			}

			if exclude {
				if matched {
					shouldAppend = false
					break
				}

			} else if matched {
				shouldAppend = true
			}

		}
		if shouldAppend {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles
}

//go:embed 	configs/language_extensions.json
var jsonData []byte

func getLanguagesExtension(languages []string) ([]string, error) {
	langMap := make(map[string]bool)
	for _, lang := range languages {
		langMap[strings.ToLower(lang)] = true
	}

	var langStructs []language
	err := json.Unmarshal(jsonData, &langStructs)
	if err != nil {
		return nil, err
	}
	for i := range langStructs {
		langStructs[i].Name = strings.ToLower(langStructs[i].Name)
	}

	extensions := make([]string, 0)
	for i := range langStructs {
		if _, ok := langMap[langStructs[i].Name]; ok {
			extensions = append(extensions, langStructs[i].Extensions...)
		}
	}
	return extensions, nil
}

type language struct {
	Name       string   `json:"name"`
	LangType   string   `json:"type"`
	Extensions []string `json:"extensions"`
}
