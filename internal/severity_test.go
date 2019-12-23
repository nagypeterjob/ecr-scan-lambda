package internal

import "testing"

var (
	a int64 = 1
	b int64 = 2
	c int64 = 3
	d int64 = 4
)

func TestCalculateScore(t *testing.T) {
	cases := []struct {
		Severity Severity
		Expected int
	}{
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL": &a,
					"HIGH":     &b,
				},
				Link: "",
			},
			Expected: 150,
		},
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL": &a,
					"HIGH":     &b,
					"MEDIUM":   &c,
				},
				Link: "",
			},
			Expected: 170,
		},
		{
			Severity: Severity{
				Count: map[string]*int64{
					"CRITICAL":  &a,
					"LOW":       &b,
					"UNDEFINED": &d,
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
