package main

import (
	"os"
	"text/template"
)

var (
	commitHash string
	versionTag string
	date       string
)

type version struct {
	CommitHash string
	Version    string
	Date       string
}

var versionTemplate = `
Commit {{ .CommitHash}}
Date:   {{ .Date}}
Version: {{ .Version}}
`

// PrintVersion outputs build related information to stdout
func printVersion() error {
	v := version{
		CommitHash: commitHash,
		Version:    versionTag,
		Date:       date,
	}

	t, err := template.New("version").Parse(versionTemplate)
	if err != nil {
		return err
	}
	err = t.Execute(os.Stdout, v)
	if err != nil {
		return err
	}
	return nil
}
