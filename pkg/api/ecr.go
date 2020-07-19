package api

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/logger"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
)

// ECRService implements ECR API
type ECRService struct {
	client     ecriface.ECRAPI
	logger     *logger.Logger
	imageTag   string
	region     string
	registryID string
}

// RepositoryInfo data structure for storing repositories
type RepositoryInfo struct {
	Name     string
	Link     string
	Severity severity.Matrix
}

// ScanningResult .
type ScanningResult struct {
	Output *ecr.PutImageScanningConfigurationOutput
	Err    error
}

// NewECRService populates a new ECRService instance
func NewECRService(registryID string, region string, imageTag string, logger *logger.Logger, client ecriface.ECRAPI) *ECRService {
	return &ECRService{
		client:     client,
		imageTag:   imageTag,
		logger:     logger,
		region:     region,
		registryID: registryID,
	}
}

func (s *ECRService) getImageScanFinding(repo *ecr.Repository) (*ecr.DescribeImageScanFindingsOutput, error) {
	describeInput := ecr.DescribeImageScanFindingsInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: aws.String(s.imageTag),
		},
		RepositoryName: repo.RepositoryName,
	}

	if len(s.registryID) != 0 {
		describeInput.RegistryId = aws.String(s.registryID)
	}
	return s.client.DescribeImageScanFindings(&describeInput)
}

func (s *ECRService) createInfo(finding *ecr.DescribeImageScanFindingsOutput) *RepositoryInfo {
	if finding.ImageScanFindings != nil && len(finding.ImageScanFindings.FindingSeverityCounts) != 0 {
		return &RepositoryInfo{
			Name: *finding.RepositoryName,
			Link: fmt.Sprintf("https://console.aws.amazon.com/ecr/repositories/%s/image/%s/scan-results?region=%s", *finding.RepositoryName, *finding.ImageId.ImageDigest, s.region),
			Severity: severity.Matrix{
				Count: finding.ImageScanFindings.FindingSeverityCounts,
			},
		}
	}
	return nil
}

func hitSeverityThreshold(info *RepositoryInfo, minimumSeverity string) bool {
	return info.Severity.CalculateScore() >= severity.SeverityTable[minimumSeverity]
}

// GatherVulnerabilities requests scan findings for repositories (for given tag)
// and filters them based on the minimum severity level.
// Returns the filtered findings and a list of repositores which couldn't be scanned.
// Reason for scanning error can be that there is no image version with the provided tag in the repsitory.
func (s *ECRService) GatherVulnerabilities(
	ctx context.Context,
	repositories chan *ecr.Repository,
	minimumSeverity string,
	numWorkers int,
) ([]*RepositoryInfo, []*RepositoryInfo) {
	var filtered []*RepositoryInfo
	var failed []*RepositoryInfo
	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for repository := range repositories {
				finding, err := s.getImageScanFinding(repository)
				if err != nil {
					mu.Lock()
					failed = append(failed, &RepositoryInfo{Name: *repository.RepositoryName})
					mu.Unlock()
				} else {
					if info := s.createInfo(finding); info != nil {
						if hitSeverityThreshold(info, minimumSeverity) {
							mu.Lock()
							filtered = append(filtered, info)
							mu.Unlock()
						}
					}
				}
			}
		}()
	}
	wg.Wait()
	return filtered, failed
}

// GenImageScanningConfiguration iterates an input repository channel
// and calls putImageScanningConfiguration on each passed repository
// then passes the result into an output channel
func (s *ECRService) GenImageScanningConfiguration(ctx context.Context, repositories chan *ecr.Repository, numWorkers int) chan *ScanningResult {
	out := make(chan *ScanningResult)
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for r := range repositories {
				select {
				case out <- s.putImageScanningConfiguration(r):
					s.logger.Infof("PutImageScanningConfiguration - (goroutine #%d)\n", i)
				case <-ctx.Done():
					s.logger.Info("genImageScanningConfiguration context cancelled")
					return
				}
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// PutImageScanningConfiguration turns on ScanOnPush for given repository
func (s *ECRService) putImageScanningConfiguration(repo *ecr.Repository) *ScanningResult {
	scanConfigInput := ecr.PutImageScanningConfigurationInput{
		RepositoryName: repo.RepositoryName,
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(true),
		},
	}
	if len(s.registryID) != 0 {
		scanConfigInput.RegistryId = aws.String(s.registryID)
	}

	o, e := s.client.PutImageScanningConfiguration(&scanConfigInput)
	return &ScanningResult{
		Output: o,
		Err:    e,
	}
}

// StartImageScan triggers image scan on given repository for a provided tag
func (s *ECRService) StartImageScan(repositoryName *string) (*ecr.StartImageScanOutput, error) {
	startImageScanInput := ecr.StartImageScanInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: aws.String(s.imageTag),
		},
		RepositoryName: repositoryName,
	}
	return s.client.StartImageScan(&startImageScanInput)
}

// DescribeRepositoriesPages iterates through all repositories and passes them into a channel
func (s *ECRService) DescribeRepositoriesPages(ctx context.Context) (chan *ecr.Repository, chan error) {
	s.logger.Info("Starting to describe repositories...")

	var wg sync.WaitGroup
	repositories := make(chan *ecr.Repository)
	errc := make(chan error, 1)

	input := &ecr.DescribeRepositoriesInput{}
	if len(s.registryID) != 0 {
		input.RegistryId = aws.String(s.registryID)
	}
	pageNum := 0
	errc <- s.client.DescribeRepositoriesPages(input, func(page *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		s.logger.Infof("Iterating repository page %d \n", pageNum)
		pageNum++
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()
			for _, output := range page.Repositories {
				select {
				case repositories <- output:
					s.logger.Infof("Describing %s has finished...(goroutine #%d)\n", *output.RepositoryName, pageNum)
				case <-ctx.Done():
					s.logger.Info("DescribeRepositoriesPages context cancelled")
					return
				}
			}
		}(pageNum)
		return true
	})
	go func() {
		wg.Wait()
		close(repositories)
		close(errc)
	}()
	return repositories, errc
}
