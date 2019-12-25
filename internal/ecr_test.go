package internal

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
)

type mockECRService struct {
	ecriface.ECRAPI
}

func (m mockECRService) DescribeRepositories(input *ecr.DescribeRepositoriesInput) (*ecr.DescribeRepositoriesOutput, error) {
	return &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{
			&ecr.Repository{
				RepositoryName: aws.String("TestRepo/Test1"),
			},
			&ecr.Repository{
				RepositoryName: aws.String("TestRepo/Test2"),
			},
			&ecr.Repository{
				RepositoryName: aws.String("TestRepo/Test3"),
			},
		},
	}, nil
}

func (m mockECRService) DescribeImageScanFindings(input *ecr.DescribeImageScanFindingsInput) (*ecr.DescribeImageScanFindingsOutput, error) {
	return &ecr.DescribeImageScanFindingsOutput{
		ImageScanFindings: &ecr.ImageScanFindings{
			FindingSeverityCounts: map[string]*int64{
				"CRITICAL": aws.Int64(12),
			},
		},
		RepositoryName: aws.String("TestRepo/Test1"),
		ImageId: &ecr.ImageIdentifier{
			ImageDigest: aws.String("xxxyyyzzzddd"),
		},
	}, nil
}

var service = NewECRService("", mockECRService{})

func TestListRepositories(t *testing.T) {
	expectedLength := 3
	result, err := service.ListRepositories(100)
	if err != nil {
		t.Fatalf("TestListRepositories failed to list repositories")
	}
	if len(result.Repositories) != expectedLength {
		t.Fatalf("TestListRepositories values are not equal, wanting: %d, got: %d", expectedLength, len(result.Repositories))
	}
}

func TestGetImageScanFinding(t *testing.T) {
	repository := ecr.Repository{
		RepositoryName: aws.String("TestRepo/Test1"),
	}
	expectedCount := int64(12)
	result, err := service.getImageScanFinding(&repository)
	severityCount := result.ImageScanFindings.FindingSeverityCounts["CRITICAL"]
	if err != nil {
		t.Fatalf("TestGetImageScanFinding failed to list repositories")
	}
	if *severityCount != expectedCount {
		t.Fatalf("TestGetImageScanFinding values are not equal, wanting: %d, got: %d", expectedCount, *severityCount)
	}
}

type inputs struct {
	finding         *ecr.DescribeImageScanFindingsOutput
	filtered        []Repository
	region          string
	minimumSeverity string
}

func TestFilterBySeverity(t *testing.T) {
	cases := []struct {
		input    inputs
		expected []Repository
	}{
		{
			// There will be 1 results as we need results above HIGH severity and we have CRITICAL
			input: inputs{
				finding: &ecr.DescribeImageScanFindingsOutput{
					ImageScanFindings: &ecr.ImageScanFindings{
						FindingSeverityCounts: map[string]*int64{
							"CRITICAL": aws.Int64(12),
						},
					},
					RepositoryName: aws.String("TestRepo/Test1"),
					ImageId: &ecr.ImageIdentifier{
						ImageDigest: aws.String("xxxyyyzzzddd"),
					},
				},
				filtered:        []Repository{},
				region:          "us-east-1",
				minimumSeverity: "HIGH",
			},
			expected: []Repository{
				Repository{
					Name: "TestRepo/Test1",
					Severity: Severity{
						Count: map[string]*int64{
							"CRITICAL": aws.Int64(12),
						},
						Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
					},
				},
			},
		},
		{
			// There will be 0 results as we need results above HIGH severity and we have LOW
			input: inputs{
				finding: &ecr.DescribeImageScanFindingsOutput{
					ImageScanFindings: &ecr.ImageScanFindings{
						FindingSeverityCounts: map[string]*int64{
							"LOW": aws.Int64(12),
						},
					},
					RepositoryName: aws.String("TestRepo/Test2"),
					ImageId: &ecr.ImageIdentifier{
						ImageDigest: aws.String("aaabbbcccddd"),
					},
				},
				filtered:        []Repository{},
				region:          "us-east-1",
				minimumSeverity: "HIGH",
			},
			expected: []Repository{},
		},
	}

	for i, c := range cases {
		repos := service.filterBySeverity(c.input.finding, c.input.filtered, c.input.region, c.input.minimumSeverity)
		if !reflect.DeepEqual(repos, c.expected) {
			t.Fatalf("[%d], values not equal, wanting: %v, got: %v", i, c.expected, repos)
		}
	}
}

func TestGetFilteredFindings(t *testing.T) {
	cases := []struct {
		input              *ecr.DescribeRepositoriesOutput
		expectedRepos      []Repository
		expectedScanErrors []ScanErrors
	}{
		{
			input: &ecr.DescribeRepositoriesOutput{
				Repositories: []*ecr.Repository{
					&ecr.Repository{
						RepositoryName: aws.String("TestRepo/Test1"),
					},
				},
			},
			expectedRepos: []Repository{
				Repository{
					Name: "TestRepo/Test1",
					Severity: Severity{
						Count: map[string]*int64{
							"CRITICAL": aws.Int64(12),
						},
						Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
					},
				},
			},
			expectedScanErrors: []ScanErrors{},
		},
	}
	for i, c := range cases {
		filtered, scanErrors := service.GetFilteredFindings(c.input, "us-east-1", "HIGH")
		if !reflect.DeepEqual(filtered, c.expectedRepos) {
			t.Fatalf("[%d], values not equal: wanting repos: %v, got: %v", i, c.expectedRepos, filtered)
		}
		if len(scanErrors) != len(c.expectedScanErrors) {
			t.Fatalf("[%d], values not equal: wanting scan errors length: %d, got: %d", i, len(c.expectedScanErrors), len(scanErrors))
		}
	}

}
