package exporters

import (
	"reflect"
	"testing"
)

func TestDomain(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{input: "ops@recart.com", expected: "recart.com"},
		{input: "", expected: ""},
		{input: "dummy@@google.com", expected: ""},
	}

	for i, c := range cases {
		m := MailgunExporter{
			from: c.input,
		}

		domain := domain(m.from)
		if !reflect.DeepEqual(c.expected, domain) {
			t.Fatalf("[%d] Error extracting domain, wanted => %s, got => %s", i, c.expected, domain)
		}
	}

}
