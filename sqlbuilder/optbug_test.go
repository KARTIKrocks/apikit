package sqlbuilder

import (
	"testing"
)

// Bug: rebasePlaceholders fast path accepts $0 but slow path does not.
// The fast path digit check is >= '0' but slow path is >= '1'.
// With a single $0 in the string, fast path fires and rebases it incorrectly.
func TestBug_RebaseFastPathAcceptsDollarZero(t *testing.T) {
	// $0 is not a valid PostgreSQL placeholder. It should be left untouched.
	// Slow path (multi-placeholder) correctly skips $0.
	slowResult := rebasePlaceholders("$0 AND $1", 5)
	// Slow path: $0 not matched (starts with '0'), $1 → $6
	expectSQL(t, "$0 AND $6", slowResult)

	// Fast path (single placeholder) should also skip $0.
	fastResult := rebasePlaceholders("x = $0", 5)
	// Should be left as "x = $0" (not "x = $5")
	expectSQL(t, "x = $0", fastResult)
}
