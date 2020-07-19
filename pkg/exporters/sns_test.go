package exporters

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
)

func TestFormatRepositories(t *testing.T) {

	input := []*api.RepositoryInfo{
		{
			Name: "TestRepository/TestRepo1",
			Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
			Severity: severity.Matrix{
				Count: map[string]*int64{
					"CRITICAL":      aws.Int64(1),
					"HIGH":          aws.Int64(2),
					"MEDIUM":        aws.Int64(3),
					"LOW":           aws.Int64(4),
					"INFORMATIONAL": aws.Int64(5),
					"UNDEFINED":     aws.Int64(6),
				},
			},
		},
		{
			Name: "TestRepository/TestRepo2",
			Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/xxxyyyzzzddd/scan-results?region=us-east-1",
			Severity: severity.Matrix{
				Count: map[string]*int64{
					"CRITICAL":  aws.Int64(6),
					"HIGH":      aws.Int64(5),
					"UNDEFINED": aws.Int64(1),
				},
			},
		},
	}

	expected := jsonData{
		Vulnerablities: []repository{
			{
				Name: "TestRepository/TestRepo1",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Findings: []vulnerablity{
					{
						Severity: "CRITICAL",
						Count:    "1",
					},
					{
						Severity: "HIGH",
						Count:    "2",
					},
					{
						Severity: "MEDIUM",
						Count:    "3",
					},
					{
						Severity: "LOW",
						Count:    "4",
					},
					{
						Severity: "INFORMATIONAL",
						Count:    "5",
					},
					{
						Severity: "UNDEFINED",
						Count:    "6",
					},
				},
			},
			{
				Name: "TestRepository/TestRepo2",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Findings: []vulnerablity{
					{
						Severity: "CRITICAL",
						Count:    "6",
					},
					{
						Severity: "HIGH",
						Count:    "5",
					},
					{
						Severity: "UNDEFINED",
						Count:    "1",
					},
				},
			},
		},
	}
	s := NewSNSExporter("sns", nil, "")

	formatted := s.format(input)
	testInput := jsonData{Vulnerablities: formatted}

	if !reflect.DeepEqual(testInput, expected) {
		t.Fatalf("Values are not equal, wanting: %s, got: %s", expected, testInput)

	}
}

func TestJsonPayload(t *testing.T) {
	input := jsonData{
		Head:    reportHeadText,
		Default: "SNS topic",
		Vulnerablities: []repository{
			{
				Name: "TestRepository/TestRepo1",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Findings: []vulnerablity{
					{
						Severity: "CRITICAL",
						Count:    "1",
					},
					{
						Severity: "HIGH",
						Count:    "2",
					},
					{
						Severity: "MEDIUM",
						Count:    "3",
					},
					{
						Severity: "LOW",
						Count:    "4",
					},
					{
						Severity: "INFORMATIONAL",
						Count:    "5",
					},
					{
						Severity: "UNDEFINED",
						Count:    "6",
					},
				},
			},
			{
				Name: "TestRepository/TestRepo2",
				Link: "https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/xxxyyyzzzddd/scan-results?region=us-east-1",
				Findings: []vulnerablity{
					{
						Severity: "CRITICAL",
						Count:    "6",
					},
					{
						Severity: "HIGH",
						Count:    "5",
					},
					{
						Severity: "UNDEFINED",
						Count:    "1",
					},
				},
			},
		},
		Failed: []string{
			"TestRepo/Failed1",
			"TestRepo/Failed2",
		},
	}

	expected := `{"head":"` + reportHeadText + `","vulnerablities":[{"name":"TestRepository/TestRepo1","link":"https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1","findings":[{"severity":"CRITICAL","count":"1"},{"severity":"HIGH","count":"2"},{"severity":"MEDIUM","count":"3"},{"severity":"LOW","count":"4"},{"severity":"INFORMATIONAL","count":"5"},{"severity":"UNDEFINED","count":"6"}]},{"name":"TestRepository/TestRepo2","link":"https://console.aws.amazon.com/ecr/repositories/TestRepo/Test2/image/xxxyyyzzzddd/scan-results?region=us-east-1","findings":[{"severity":"CRITICAL","count":"6"},{"severity":"HIGH","count":"5"},{"severity":"UNDEFINED","count":"1"}]}],"failed":["TestRepo/Failed1","TestRepo/Failed2"],"default":"SNS topic"}`

	js, err := marshal(input)
	if err != nil {
		t.Fatalf("Error marshaling json %s", err)
	}
	if !reflect.DeepEqual(string(js), expected) {
		t.Fatalf("Marshalled json is not what I expected, wanted => %s, got => %s", expected, js)
	}
}
