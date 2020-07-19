package exporters

import (
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
)

// Exporter defines a common interface for different exporters
type Exporter interface {
	// Formats message types then returns function which sends formatted messages on invocation
	Format(filtered []*api.RepositoryInfo, failed []*api.RepositoryInfo) (func() error, error)
	// Retrun exporter name
	Name() string
}
