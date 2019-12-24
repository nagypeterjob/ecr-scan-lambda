package internal

import (
	"os"
	"text/template"
)

var (
	commitHash string
	versionTag string
	date       string
	author     string
)

type version struct {
	CommitHash string
	Version    string
	Date       string
	Author     string
}

var versionTemplate = `
Commit {{ .CommitHash}}
Author: {{ .Author }}
Date:   {{ .Date}}
Version: {{ .Version}}
`

// PrintVersion outputs build related information to stdout
func PrintVersion() error {
	v := version{
		CommitHash: commitHash,
		Version:    versionTag,
		Date:       date,
		Author:     author,
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
