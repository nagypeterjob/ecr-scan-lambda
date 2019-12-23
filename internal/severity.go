package internal

type Severity struct {
	Count map[string]*int64
	Link  string
}

type Repository struct {
	Name     string
	Severity Severity
}

type ScanErrors struct {
	RepositoryName string
}

var SeverityList = []string{
	"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFORMATIONAL", "UNDEFINED",
}

var SeverityTable = map[string]int{
	"CRITICAL":      CriticalSeverityScore,
	"HIGH":          HighSeverityScore,
	"MEDIUM":        MediumSeverityScore,
	"LOW":           LowSeverityScore,
	"INFORMATIONAL": InformationalSeverityScore,
	"UNDEFINED":     UndefinedSeverityScore,
}

const (
	CriticalSeverityScore      int = 100
	HighSeverityScore          int = 50
	MediumSeverityScore        int = 20
	LowSeverityScore           int = 10
	InformationalSeverityScore int = 5
	UndefinedSeverityScore     int = 1
)

func (sev *Severity) CalculateScore() int {
	score := 0
	for k, _ := range sev.Count {
		score += SeverityTable[k]
	}
	return score
}
