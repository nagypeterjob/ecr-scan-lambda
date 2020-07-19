package api

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/logger"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
)

type mockECRService struct {
	ecriface.ECRAPI
}

func (m mockECRService) DescribeImageScanFindings(input *ecr.DescribeImageScanFindingsInput) (*ecr.DescribeImageScanFindingsOutput, error) {
	switch *input.RepositoryName {
	case "TestRepo/Test1":
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
	case "TestRepo/Test2":
		return &ecr.DescribeImageScanFindingsOutput{
			ImageScanFindings: &ecr.ImageScanFindings{
				FindingSeverityCounts: map[string]*int64{
					"HIGH": aws.Int64(32),
					"LOW":  aws.Int64(1),
				},
			},
			RepositoryName: aws.String("TestRepo/Test2"),
			ImageId: &ecr.ImageIdentifier{
				ImageDigest: aws.String("aaabbbccddd"),
			},
		}, nil
	case "TestRepo/Test3":
		return &ecr.DescribeImageScanFindingsOutput{
			ImageScanFindings: &ecr.ImageScanFindings{
				FindingSeverityCounts: map[string]*int64{
					"LOW": aws.Int64(1),
				},
			},
			RepositoryName: aws.String("TestRepo/Test3"),
			ImageId: &ecr.ImageIdentifier{
				ImageDigest: aws.String("eeefffggghhh"),
			},
		}, nil
	default:
		return &ecr.DescribeImageScanFindingsOutput{
			ImageScanFindings: &ecr.ImageScanFindings{
				FindingSeverityCounts: nil,
			},
			RepositoryName: aws.String("TestRepo/Test4"),
			ImageId: &ecr.ImageIdentifier{
				ImageDigest: aws.String("mmmnnnoooppp"),
			},
		}, fmt.Errorf("Fake error happened")
	}
}

var service *ECRService

func init() {

	logger, err := logger.NewLogger("DEBUG")
	if err != nil {
		panic(err)
	}
	service = NewECRService("xxxxx", "us-east-1", "latest", logger, mockECRService{})
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
	finding *ecr.DescribeImageScanFindingsOutput
	region  string
}

func TestCreateInfo(t *testing.T) {
	cases := []struct {
		input    inputs
		expected *RepositoryInfo
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
				region: "us-east-1",
			},
			expected: &RepositoryInfo{
				Name: "TestRepo/Test1",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"CRITICAL": aws.Int64(12),
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
				region: "us-east-1",
			},
			expected: &RepositoryInfo{
				Name: "TestRepo/Test2",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/aaabbbcccddd/scan-results?region=us-east-1",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"LOW": aws.Int64(12),
					},
				},
			},
		},
	}

	for i, c := range cases {
		repos := service.createInfo(c.input.finding)
		if !reflect.DeepEqual(repos, c.expected) {
			t.Fatalf("[%d], values not equal, wanting: %v, got: %v", i, c.expected, repos)
		}
	}
}

func TestHitSeverityThreshold(t *testing.T) {
	cases := []struct {
		input    *RepositoryInfo
		expected bool
	}{
		{
			input: &RepositoryInfo{
				Name: "TestRepo/Test1",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"LOW": aws.Int64(12),
					},
				},
			},
			expected: false,
		},
		{
			input: &RepositoryInfo{
				Name: "TestRepo/Test2",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/aaabbbcccddd/scan-results?region=us-east-1",
				Severity: severity.Matrix{
					Count: map[string]*int64{
						"HIGH": aws.Int64(10),
					},
				},
			},
			expected: true,
		},
	}
	for i, c := range cases {
		repos := hitSeverityThreshold(c.input, "MEDIUM")
		if !reflect.DeepEqual(repos, c.expected) {
			t.Fatalf("[%d], values not equal, wanting: %v, got: %v", i, c.expected, repos)
		}
	}
}

func gen(repositories []*ecr.Repository) chan *ecr.Repository {
	ret := make(chan *ecr.Repository)
	go func() {
		for _, r := range repositories {
			ret <- r
		}
		close(ret)
	}()
	return ret
}

func contains(str string, arr []string) bool {
	for _, obj := range arr {
		if obj == str {
			return true
		}
	}
	return false
}

func TestGatherVulnerabilities(t *testing.T) {
	repositories := []*ecr.Repository{
		{
			RepositoryName: aws.String("TestRepo/Test1"),
		},
		{
			RepositoryName: aws.String("TestRepo/Test2"),
		},
		{
			RepositoryName: aws.String("TestRepo/Test3"),
		},
		{
			RepositoryName: aws.String("TestRepo/NoVulnerablity"),
		},
	}

	input := gen(repositories)
	expectedFiltered := []string{
		"TestRepo/Test1",
		"TestRepo/Test2",
	}
	expectedFailed := []string{
		"TestRepo/NoVulnerablity",
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	filtered, failed := service.GatherVulnerabilities(ctx, input, "MEDIUM", 2)

	for _, f := range filtered {
		if !contains(f.Name, expectedFiltered) {
			t.Errorf("Filtered expected to contain %s", f.Name)
		}
	}

	for _, f := range failed {
		if !contains(f.Name, expectedFailed) {
			t.Fatalf("Failed expected to contain %s", f.Name)
		}
	}
}

func (m mockECRService) PutImageScanningConfiguration(input *ecr.PutImageScanningConfigurationInput) (*ecr.PutImageScanningConfigurationOutput, error) {
	return &ecr.PutImageScanningConfigurationOutput{
		RepositoryName: aws.String(*input.RepositoryName),
	}, nil
}

func TestGenImageScanningConfiguration(t *testing.T) {
	repositories := []*ecr.Repository{
		{
			RepositoryName: aws.String("TestRepo/Test1"),
		},
		{
			RepositoryName: aws.String("TestRepo/Test2"),
		},
		{
			RepositoryName: aws.String("TestRepo/Test3"),
		},
		{
			RepositoryName: aws.String("TestRepo/NoVulnerablity"),
		},
	}

	expected := []ScanningResult{
		{
			Output: &ecr.PutImageScanningConfigurationOutput{
				RepositoryName: aws.String("TestRepo/Test1"),
			},
			Err: nil,
		},
		{
			Output: &ecr.PutImageScanningConfigurationOutput{
				RepositoryName: aws.String("TestRepo/Test2"),
			},
			Err: nil,
		},
		{
			Output: &ecr.PutImageScanningConfigurationOutput{
				RepositoryName: aws.String("TestRepo/Test3"),
			},
			Err: nil,
		},
		{
			Output: &ecr.PutImageScanningConfigurationOutput{
				RepositoryName: aws.String("TestRepo/NoVulnerablity"),
			},
			Err: nil,
		},
	}

	input := gen(repositories)
	scanrResults := service.GenImageScanningConfiguration(context.Background(), input, 2)

	var results []ScanningResult
	for s := range scanrResults {
		results = append(results, *s)
	}
	for i, e := range expected {
		if e.Output.RepositoryName == results[i].Output.RepositoryName {
			t.Errorf("[%d] Error in TesGenImageScanningConfiguration, wanted => %v, got => %v", i, e, results[i])
		}
	}
}

func (m mockECRService) DescribeRepositoriesPages(input *ecr.DescribeRepositoriesInput, fn func(*ecr.DescribeRepositoriesOutput, bool) bool) error {
	output := &ecr.DescribeRepositoriesOutput{
		Repositories: []*ecr.Repository{
			{
				RepositoryName: aws.String("TestRepo/Test1"),
			},
			{
				RepositoryName: aws.String("TestRepo/Test2"),
			},
			{
				RepositoryName: aws.String("TestRepo/Test3"),
			},
		},
	}
	fn(output, true)
	return nil
}

func TestDescribeRepositoriesPages(t *testing.T) {
	expectedLen := 3
	cnt := 0

	repositories, errc := service.DescribeRepositoriesPages(context.Background())

	for range repositories {
		cnt++
	}

	if err := <-errc; err != nil {
		t.Fatalf("TestDescribeRepositoriesPages failed to list repositories")
	}

	if cnt != expectedLen {
		t.Fatalf("TestDescribeRepositoriesPages values are not equal, wanting: %d, got: %d", expectedLen, cnt)
	}
}
