package exporters

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/api"
	"github.com/nagypeterjob/ecr-scan-lambda/pkg/severity"
)

func TestTmpl(t *testing.T) {
	expected := `Vulnerabilities found in TestRepo/Test1:`

	data := values{
		Name: "TestRepo/Test1",
	}

	head, err := execTmpl(data, `Vulnerabilities found in {{ .Name }}:`)
	if err != nil {
		t.Fatalf("Runtime error formatting text: %s", err)
	}

	if !reflect.DeepEqual(head.String(), expected) {
		t.Fatalf("Error formatting text => wanted: \n%s, got: \n%s", expected, head.String())
	}
}

var input = api.RepositoryInfo{
	Name: "TestRepo/Test1",
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
}

func TestTextFormatSingle(t *testing.T) {

	expectedMsg := `Vulnerabilities found in TestRepo/Test1:

     CRITICAL: 1
         HIGH: 2
       MEDIUM: 3
          LOW: 4
INFORMATIONAL: 5
    UNDEFINED: 6

View detailed scan results on console (https://console.aws.amazon.com/ecr/repositories/TestRepo/Test1/image/xxxyyyzzzddd/scan-results?region=us-east-1)
--------------------------------------
`

	msg, err := fillTmpl(&input)
	if err != nil {
		t.Fatalf("Runtime error `formatt`ing text: %s", err)
	}

	if !reflect.DeepEqual(expectedMsg, msg) {
		t.Fatalf("Error `formatt`ing text => wanted: \n%v, got: \n%v", expectedMsg, msg)
	}
	if !reflect.DeepEqual(expectedMsg, msg) {
		t.Fatalf("Error `formatt`ing text => wanted: \n%v, got: \n%v", expectedMsg, msg)
	}
	if !reflect.DeepEqual(expectedMsg, msg) {
		t.Fatalf("Error `formatt`ing text => wanted: \n%v, got: \n%v", expectedMsg, msg)
	}
}
