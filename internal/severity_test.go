package internal

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestCalculateScore(t *testing.T) {
	cases := []struct {
		Severity Severity
		Expected int
	}{
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL": aws.Int64(1),
					"HIGH":     aws.Int64(2),
				},
				Link: "",
			},
			Expected: 150,
		},
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL": aws.Int64(1),
					"HIGH":     aws.Int64(2),
					"MEDIUM":   aws.Int64(3),
				},
				Link: "",
			},
			Expected: 170,
		},
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL":  aws.Int64(1),
					"LOW":       aws.Int64(2),
					"UNDEFINED": aws.Int64(3),
				},
				Link: "",
			},
			Expected: 111,
		},
	}

	for i, c := range cases {
		score := c.Severity.CalculateScore()
		if score != c.Expected {
			t.Fatalf("[%d], values not equal, wanting: %d, got: %d", i, c.Expected, score)
		}
	}
}
