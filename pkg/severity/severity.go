package severity

// Matrix data structure for storing severity
type Matrix struct {
	Count map[string]*int64
}

// SeverityList holds severity levels and maintains order during iteration
var SeverityList = []string{
	"CRITICAL", "HIGH", "MEDIUM", "LOW", "INFORMATIONAL", "UNDEFINED",
}

// SeverityTable maps a score to each severity level
var SeverityTable = map[string]int{
	"CRITICAL":      criticalSeverityScore,
	"HIGH":          highSeverityScore,
	"MEDIUM":        mediumSeverityScore,
	"LOW":           lowSeverityScore,
	"INFORMATIONAL": informationalSeverityScore,
	"UNDEFINED":     undefinedSeverityScore,
}

const (
	criticalSeverityScore      int = 100
	highSeverityScore          int = 50
	mediumSeverityScore        int = 20
	lowSeverityScore           int = 10
	informationalSeverityScore int = 5
	undefinedSeverityScore     int = 1
)

// CalculateScore calculates severity score to each finding
func (sev *Matrix) CalculateScore() int {
	score := 0
	for k := range sev.Count {
		score += SeverityTable[k]
	}
	return score
}
