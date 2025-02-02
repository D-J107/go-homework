package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func checkFlags(repository, revision, orderBy, format string) error {
	cmd := exec.Command("git", "ls-tree", "-r", revision)
	cmd.Dir = repository
	var outBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git ls-tree revision errro: %v", err)
	}
	if !strings.Contains("lines commits files", orderBy) {
		return errors.New("unknown --order-by flag value")
	}
	if !strings.Contains("tabular csv json json-lines", format) {
		return errors.New("unknown --format flag value")
	}
	return nil
}
