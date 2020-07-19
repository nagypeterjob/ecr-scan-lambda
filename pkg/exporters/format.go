package exporters

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
)

type values struct {
	Name               string
	CountCritical      *int64
	CountHigh          *int64
	CountMedium        *int64
	CountLow           *int64
	CountInformational *int64
	CountUndefined     *int64
	Link               string
}

var (
	// Vulnerablity list header
	reportHeadText = fmt.Sprintf("Scan results on %s", time.Now().Format("2006 Jan 02"))
	// Failed scan list header
	reportFailedHeadText = "Failed to get scan results from the following repos:"
	// Message in case no vulnerablity hit the threshold
	reportClean = "Looks like the tested images have zero vulnerabilities hitting the threshold, good job!"
)

func execTmpl(data interface{}, raw string) (bytes.Buffer, error) {
	tmpl, err := template.New("text formatter").Parse(raw)
	if err != nil {
		return bytes.Buffer{}, err
	}
	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, &data)
	if err != nil {
		return bytes.Buffer{}, err
	}
	return buffer, nil
}

func fillTmpl(r *api.RepositoryInfo) (string, error) {

	data := values{
		Name:               r.Name,
		CountCritical:      r.Severity.Count["CRITICAL"],
		CountHigh:          r.Severity.Count["HIGH"],
		CountMedium:        r.Severity.Count["MEDIUM"],
		CountLow:           r.Severity.Count["LOW"],
		CountInformational: r.Severity.Count["INFORMATIONAL"],
		CountUndefined:     r.Severity.Count["UNDEFINED"],
		Link:               r.Link,
	}

	raw := `Vulnerabilities found in {{ .Name }}:
{{printf "%s" "\n"}}
{{- if .CountCritical }}     CRITICAL: {{ .CountCritical }}{{printf "%s" "\n"}}{{end}}
{{- if .CountHigh }}         HIGH: {{ .CountHigh }}{{printf "%s" "\n"}}{{end}}
{{- if .CountMedium }}       MEDIUM: {{ .CountMedium }}{{printf "%s" "\n"}}{{end}}
{{- if .CountLow }}          LOW: {{ .CountLow }}{{printf "%s" "\n"}}{{end}}
{{- if .CountInformational }}INFORMATIONAL: {{ .CountInformational }}{{printf "%s" "\n"}}{{end}}
{{- if .CountUndefined }}    UNDEFINED: {{ .CountUndefined }}{{end}}

View detailed scan results on console ({{ .Link }})
--------------------------------------
`

	str, err := execTmpl(data, raw)
	if err != nil {
		return "", err
	}

	return str.String(), err
}

// formats concatenates textual representation of vulnerablities to one string
func format(repositories []*api.RepositoryInfo) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(reportHeadText + "\n")

	if len(repositories) == 0 {
		buffer.WriteString(reportClean + "\n")
		return buffer.String(), nil
	}

	for _, r := range repositories {
		msg, err := fillTmpl(r)
		if err != nil {
			return "", err
		}
		buffer.WriteString(msg)
	}
	return buffer.String(), nil
}

// formatFailed creates a list of repositories which has failed scanning
func formatFailed(repositories []*api.RepositoryInfo) (string, error) {
	var buffer bytes.Buffer
	if len(repositories) == 0 {
		return "", nil
	}

	buffer.WriteString(reportFailedHeadText + "\n")
	for _, r := range repositories {
		buffer.WriteString(r.Name + "\n")
	}
	return buffer.String(), nil
}
