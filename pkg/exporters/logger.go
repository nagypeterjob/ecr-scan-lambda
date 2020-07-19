package exporters

import (
	"fmt"

	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
)

// LogExporter is a dummy exporter which prints formatted message to stdout.
// For debug and educational purposes.
type LogExporter struct {
	name string
}

// NewLogExporter .
func NewLogExporter(name string) *LogExporter {
	return &LogExporter{
		name: name,
	}
}

// Name .
func (l LogExporter) Name() string {
	return l.name
}

// Format clousure formats scan results and returns a function that sends report on invocation
func (l LogExporter) Format(filtered []*api.RepositoryInfo, failed []*api.RepositoryInfo) (func() error, error) {
	filteredMsg, err := format(filtered)
	if err != nil {
		return nil, err
	}

	failedMsg, err := formatFailed(failed)
	if err != nil {
		return nil, err
	}

	msg := filteredMsg + failedMsg

	return func() error {
		fmt.Println(msg)
		return nil
	}, nil
}
