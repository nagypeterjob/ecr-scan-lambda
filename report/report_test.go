package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/nagypeterjob/ecr-scan-lambda/internal"
)

type mockApp struct{}

type mockECR struct {
	ecriface.ECRAPI
}

func (a *mockApp) listRepositories(maxRepos int) (*ecr.DescribeRepositoriesOutput, error) {
	return &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{
			{RepositoryName: aws.String("test/repository1")},
			{RepositoryName: aws.String("test/repository2")},
			{RepositoryName: aws.String("test/repository3")},
		},
	}, nil
}

func (a mockECR) DescribeImageScanFindings(input *ecr.DescribeImageScanFindingsInput) (*ecr.DescribeImageScanFindingsOutput, error) {
	if *input.RepositoryName == "test/repository1" {
		return &ecr.DescribeImageScanFindingsOutput{}, errors.New("Failed to get ImageScanFindings")
	}

	if *input.RepositoryName == "test/repository2" {
		return &ecr.DescribeImageScanFindingsOutput{
			RepositoryName: input.RepositoryName,
			ImageId: &ecr.ImageIdentifier{
				ImageDigest: aws.String("aaabbbcccddd"),
			},
			ImageScanFindings: &ecr.ImageScanFindings{
				FindingSeverityCounts: map[string]*int64{
					"HIGH": aws.Int64(4),
					"LOW":  aws.Int64(2),
				},
			},
		}, nil
	}

	return &ecr.DescribeImageScanFindingsOutput{
		RepositoryName: input.RepositoryName,
		ImageId: &ecr.ImageIdentifier{
			ImageDigest: aws.String("xxxyyyzzzddd"),
		},
		ImageScanFindings: &ecr.ImageScanFindings{
			FindingSeverityCounts: map[string]*int64{
				"CRITICAL": aws.Int64(1),
				"HIGH":     aws.Int64(2),
			},
		},
	}, nil
}

var mockAppInstance = mockApp{}
var appInstance = app{
	ecrService:      mockECR{},
	region:          "us-east-1",
	minimumSeverity: "CRITICAL",
}

func (a *mockApp) GetFindings(r *ecr.DescribeRepositoriesOutput) ([]ecr.DescribeImageScanFindingsOutput, []internal.ScanErrors) {
	return appInstance.GetFindings(r)
}

func (a *mockApp) filterBySeverity(findings []ecr.DescribeImageScanFindingsOutput) []internal.Repository {
	return appInstance.filterBySeverity(findings)
}

func TestGetFindings(t *testing.T) {
	expectedFindings := 2
	expectedErrors := 1

	list, _ := mockAppInstance.listRepositories(1000)
	findings, scanErrors := mockAppInstance.GetFindings(list)
	if len(findings) != expectedFindings || len(scanErrors) != expectedErrors {
		t.Fatalf("TestGetFindings failed, values not equal, wanting findings to be: %d, got: %d and scanErrors to be:  %d, got: %d", expectedFindings, len(findings), expectedErrors, len(scanErrors))
	}
}

func TestFilterBySeverity(t *testing.T) {
	expected := []internal.Repository{
		{
			Name: "test/repository3",
			Severity: internal.Severity{
				Count: map[string]*int64{
					"CRITICAL": aws.Int64(1),
					"HIGH":     aws.Int64(2),
				},
				Link: "https://console.aws.amazon.com/ecr/repositories/test/repository3/image/xxxyyyzzzddd/scan-results?region=us-east-1",
			},
		},
	}
	list, _ := mockAppInstance.listRepositories(1000)
	findings, _ := mockAppInstance.GetFindings(list)

	filtered := mockAppInstance.filterBySeverity(findings)
	if !reflect.DeepEqual(expected, filtered) {
		t.Fatalf("Values not equal, wanting findings to be: %v, got: %v", expected, filtered)

	}
}
