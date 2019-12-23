package internal

import (
	"os"
	"text/template"
)

var (
	CommitHash string
	Version    string
	Date       string
	Author     string
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

func PrintVersion() error {
	v := version{
		CommitHash: CommitHash,
		Version:    Version,
		Date:       Date,
		Author:     Author,
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
