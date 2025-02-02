//go:build !solution

package main

import (
	"github.com/spf13/pflag"
	"os"
)

func main() {

	flagRepository := pflag.String("repository", ".", "path to repository")
	flagRevision := pflag.String("revision", "HEAD", "commit pointer")
	flagOrderBy := pflag.String("order-by", "lines", "order by sort key")
	flagUseCommitter := pflag.Bool("use-committer", false, "replace author to commiter")
	flagFormat := pflag.String("format", "tabular", "output format")
	flagExtensions := pflag.StringSlice("extensions", nil, "extensions to search")
	flagLanguages := pflag.StringSlice("languages", nil, "languages to search")
	flagExclude := pflag.StringSlice("exclude", nil, "Glob patterns to exclude")
	flagRestrictTo := pflag.StringSlice("restrict-to", nil, "Glob patterns to restrict")
	flagCPUCount := pflag.Int("cpu-count", 16, "number of CPUs to use for concurrency")

	pflag.Parse()

	if err := checkFlags(*flagRepository, *flagRevision, *flagOrderBy, *flagFormat); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	filesToCalculateStatistics, _ := findMatchFiles(
		*flagRepository, *flagRevision, *flagExtensions,
		*flagLanguages, *flagExclude, *flagRestrictTo)

	authorMap := CalculateStatisticsForFile(*flagCPUCount,
		filesToCalculateStatistics, *flagRepository, *flagRevision, *flagUseCommitter)

	arr := getSortedArray(authorMap, *flagOrderBy)
	err := prettyPrint(arr, *flagFormat)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

}
