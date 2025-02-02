package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"text/tabwriter"
)

func prettyPrint(arr []sortedInfo, format string) error {

	var w *tabwriter.Writer
	if format == "tabular" {
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	}
	if format == "csv" {
		w = tabwriter.NewWriter(os.Stdout, 0, 4, 1, '\t', 0)
	}

	outputFormat := formFormat(format)
	header := formHeaderLine(format)
	if header != "" {
		_, err := fmt.Fprintln(w, header)
		if err != nil {
			return err
		}
	}

	if outputFormat == "json" {
		output := make([]sortedInfo, 0)
		for i := range arr {
			output = append(output, sortedInfo{
				Name:    arr[i].Name,
				Lines:   arr[i].Lines,
				Commits: arr[i].Commits,
				Files:   arr[i].Files,
			})
		}
		bytes, err := json.Marshal(output)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
		return nil
	}

	if outputFormat == "json-lines" {
		for i := range arr {
			output := sortedInfo{Name: arr[i].Name, Lines: arr[i].Lines,
				Files: arr[i].Files, Commits: arr[i].Commits}
			bytes, err := json.Marshal(output)
			if err != nil {
				return err
			}
			fmt.Println(string(bytes))
		}
		return nil
	}

	for i := range arr {
		_, err := fmt.Fprintf(w, outputFormat,
			arr[i].Name, arr[i].Lines, arr[i].Commits, arr[i].Files,
		)
		if err != nil {
			return err
		}
	}

	var err error
	if w != nil {
		err = w.Flush()
	}
	if err != nil {
		return err
	}
	return nil
}

func getSortedArray(authorMap map[string]programmerInfo, orderBy string) []sortedInfo {
	output := make([]sortedInfo, 0)
	for name := range authorMap {
		info := sortedInfo{
			Name:    name,
			Lines:   authorMap[name].lines,
			Commits: len(authorMap[name].commits),
			Files:   len(authorMap[name].files),
		}
		output = append(output, info)
	}

	switch orderBy {
	case "lines":
		slices.SortFunc(output, func(a, b sortedInfo) int {
			return -1 * cmp.Or(
				cmp.Compare(a.Lines, b.Lines),
				cmp.Compare(a.Commits, b.Commits),
				cmp.Compare(a.Files, b.Files),
				cmp.Compare(b.Name, a.Name),
			)
		})
	case "commits":
		slices.SortFunc(output, func(a, b sortedInfo) int {
			return -1 * cmp.Or(
				cmp.Compare(a.Commits, b.Commits),
				cmp.Compare(a.Lines, b.Lines),
				cmp.Compare(a.Files, b.Files),
				cmp.Compare(b.Name, a.Name),
			)
		})
	case "files":
		slices.SortFunc(output, func(a, b sortedInfo) int {
			return -1 * cmp.Or(
				cmp.Compare(a.Files, b.Files),
				cmp.Compare(a.Lines, b.Lines),
				cmp.Compare(a.Commits, b.Commits),
				cmp.Compare(b.Name, a.Name),
			)
		})
	}

	return output
}

type sortedInfo struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

func formHeaderLine(format string) string {
	switch format {
	case "tabular":
		return "Name\tLines\tCommits\tFiles"
	case "csv":
		return "Name,Lines,Commits,Files"
	case "json":
		return ""
	case "json-lines":
		return ""
	default:
		return "unknown"
	}
}

func formFormat(format string) string {
	switch format {
	case "tabular":
		return "%s\t%d\t%d\t%d\n"
	case "csv":
		return "%s,%d,%d,%d\n"
	case "json":
		return "json"
	case "json-lines":
		return "json-lines"
	default:
		return "unknown"
	}
}

//
//func syncMapLen(m *sync.Map) int {
//	i := 0
//	m.Range(func(key, value interface{}) bool {
//		i++
//		return true
//	})
//	return i
//}
