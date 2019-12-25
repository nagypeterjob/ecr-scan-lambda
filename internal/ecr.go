package internal

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

// ECRService implements ECR API
type ECRService struct {
	client     ecriface.ECRAPI
	registryID string
}

// NewECRService populates a new ECRService instance
func NewECRService(registryID string, client ecriface.ECRAPI) *ECRService {
	return &ECRService{
		client:     client,
		registryID: registryID,
	}
}

//ListRepositories describes and returns ecr repositories
func (svc *ECRService) ListRepositories(maxRepos int) (*ecr.DescribeRepositoriesOutput, error) {
	mr := int64(maxRepos)
	input := ecr.DescribeRepositoriesInput{
		MaxResults: &mr,
	}

	if len(svc.registryID) != 0 {
		input.RegistryId = aws.String(svc.registryID)
	}

	return svc.client.DescribeRepositories(&input)
}

func (svc *ECRService) getImageScanFinding(repo *ecr.Repository) (*ecr.DescribeImageScanFindingsOutput, error) {
	describeInput := ecr.DescribeImageScanFindingsInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: aws.String("latest"),
		},
		RepositoryName: repo.RepositoryName,
	}

	if len(svc.registryID) != 0 {
		describeInput.RegistryId = aws.String(svc.registryID)
	}
	return svc.client.DescribeImageScanFindings(&describeInput)
}

func (svc *ECRService) filterBySeverity(finding *ecr.DescribeImageScanFindingsOutput, filtered []Repository, region string, minimumSeverity string) []Repository {
	if finding.ImageScanFindings != nil && len(finding.ImageScanFindings.FindingSeverityCounts) != 0 {
		r := Repository{
			Name: *finding.RepositoryName,
			Severity: Severity{
				Count: finding.ImageScanFindings.FindingSeverityCounts,
				Link:  fmt.Sprintf("https://console.aws.amazon.com/ecr/repositories/%s/image/%s/scan-results?region=%s", *finding.RepositoryName, *finding.ImageId.ImageDigest, region),
			},
		}
		if r.Severity.CalculateScore() >= SeverityTable[minimumSeverity] {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// GetFilteredFindings requests scan findings for repositories and filters them based on given severity level.
// It returns the filtered findings and a list of repositores which couldn't be scanned.
// Reason for scanning error can be that there is no image version with latest tag in the repsitory.
func (svc *ECRService) GetFilteredFindings(r *ecr.DescribeRepositoriesOutput, region string, minimumSeverity string) ([]Repository, []ScanErrors) {
	var filtered []Repository
	var failed []ScanErrors

	for _, repo := range r.Repositories {
		finding, err := svc.getImageScanFinding(repo)
		if err != nil {
			failed = append(failed, ScanErrors{RepositoryName: *repo.RepositoryName})
		} else {
			filtered = svc.filterBySeverity(finding, filtered, region, minimumSeverity)
		}
	}
	return filtered, failed
}

// PutImageScanningConfiguration turns on ScanOnPush for given repository
func (svc *ECRService) PutImageScanningConfiguration(repo *ecr.Repository) (*ecr.PutImageScanningConfigurationOutput, error) {
	scanConfigInput := ecr.PutImageScanningConfigurationInput{
		RepositoryName: repo.RepositoryName,
		ImageScanningConfiguration: &ecr.ImageScanningConfiguration{
			ScanOnPush: aws.Bool(true),
		},
	}
	if len(svc.registryID) != 0 {
		scanConfigInput.RegistryId = aws.String(svc.registryID)
	}

	return svc.client.PutImageScanningConfiguration(&scanConfigInput)
}

// StartImageScan runs image scan on given repository's latest image
func (svc *ECRService) StartImageScan(repositoryName *string) (*ecr.StartImageScanOutput, error) {
	startImageScanInput := ecr.StartImageScanInput{
		ImageId: &ecr.ImageIdentifier{
			ImageTag: aws.String("latest"),
		},
		RepositoryName: repositoryName,
	}
	return svc.client.StartImageScan(&startImageScanInput)
}
