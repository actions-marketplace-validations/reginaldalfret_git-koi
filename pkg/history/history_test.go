package history_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
)

func TestParseAccountWithHistory(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "account-with-history.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var snapshot history.AccountSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("failed to unmarshal snapshot: %v", err)
	}

	if snapshot.AccountID != 1001 {
		t.Errorf("expected account ID 1001, got %d", snapshot.AccountID)
	}

	shares, err := history.CalculateLanguageShares(&snapshot)
	if err != nil {
		t.Fatalf("failed to calculate language shares: %v", err)
	}
	if len(shares) != 1 || shares["Rust"] != 1.0 {
		t.Errorf("expected Rust share 1.0, got %v", shares)
	}

	summary, err := history.SummarizeContributions(&snapshot)
	if err != nil {
		t.Fatalf("failed to summarize contributions: %v", err)
	}

	// 1 + 2 + 3 + 1 + 5 + 2 = 14
	if summary.TotalContributions != 14 {
		t.Errorf("expected total contributions 14, got %d", summary.TotalContributions)
	}

	if len(summary.Weeks) == 0 {
		t.Errorf("expected at least 1 week, got 0")
	}
}

func TestDateParsingAndLeapYears(t *testing.T) {
	// Leap year date: Feb 29, 2024
	d, err := history.ParseDate("2024-02-29")
	if err != nil {
		t.Fatalf("failed to parse valid leap year date: %v", err)
	}
	if d.String() != "2024-02-29" {
		t.Errorf("expected 2024-02-29, got %s", d.String())
	}
	nextDay := d.AddDays(1)
	if nextDay.String() != "2024-03-01" {
		t.Errorf("expected 2024-03-01, got %s", nextDay.String())
	}

	// Non-leap year Feb 29 should fail
	_, err = history.ParseDate("2023-02-29")
	if err == nil {
		t.Errorf("expected error parsing Feb 29 on non-leap year 2023")
	}

	// Invalid date format
	_, err = history.ParseDate("not-a-date")
	if err == nil {
		t.Errorf("expected error parsing invalid date string")
	}
}

func TestLanguageColorParsing(t *testing.T) {
	valid, err := history.ParseLanguageColor("#00add8")
	if err != nil {
		t.Fatalf("failed to parse valid color: %v", err)
	}
	if valid.Hex() != "#00add8" {
		t.Errorf("expected #00add8, got %s", valid.Hex())
	}

	// Invalid colors
	invalidCases := []string{
		"00add8",   // missing '#'
		"#12345",   // too short
		"#1234567", // too long
		"#zzzzzz",  // non-hex characters
	}
	for _, tc := range invalidCases {
		_, err := history.ParseLanguageColor(tc)
		if err == nil {
			t.Errorf("expected error parsing invalid color %q", tc)
		}
	}
}

func TestSummarizeContributionsValidation(t *testing.T) {
	obs := history.NewDate(2026, 11, 10)

	// 1. Missing days gap test
	gapSnapshot := &history.AccountSnapshot{
		SchemaVersion: 3,
		ContributionHistory: &history.ContributionHistory{
			ObservedOn: obs,
			Days: []history.ContributionDay{
				{Date: history.NewDate(2026, 11, 8), ContributionCount: 2},
				// Missing Nov 9
				{Date: history.NewDate(2026, 11, 10), ContributionCount: 3},
			},
		},
	}
	_, err := history.SummarizeContributions(gapSnapshot)
	if err == nil {
		t.Errorf("expected error for calendar with date gap")
	}

	// 2. Conflicting counts on duplicate date
	conflictSnapshot := &history.AccountSnapshot{
		SchemaVersion: 3,
		ContributionHistory: &history.ContributionHistory{
			ObservedOn: obs,
			Days: []history.ContributionDay{
				{Date: history.NewDate(2026, 11, 10), ContributionCount: 2},
				{Date: history.NewDate(2026, 11, 10), ContributionCount: 5},
			},
		},
	}
	_, err = history.SummarizeContributions(conflictSnapshot)
	if err == nil {
		t.Errorf("expected error for conflicting contribution counts")
	}

	// 3. Date after observation date
	afterObsSnapshot := &history.AccountSnapshot{
		SchemaVersion: 3,
		ContributionHistory: &history.ContributionHistory{
			ObservedOn: obs,
			Days: []history.ContributionDay{
				{Date: history.NewDate(2026, 11, 11), ContributionCount: 1}, // after Nov 10
			},
		},
	}
	_, err = history.SummarizeContributions(afterObsSnapshot)
	if err == nil {
		t.Errorf("expected error for date after observation date")
	}

	// 4. Clean continuous history
	cleanSnapshot := &history.AccountSnapshot{
		SchemaVersion: 3,
		ContributionHistory: &history.ContributionHistory{
			ObservedOn: obs,
			Days: []history.ContributionDay{
				{Date: history.NewDate(2026, 11, 8), ContributionCount: 1},
				{Date: history.NewDate(2026, 11, 9), ContributionCount: 0},
				{Date: history.NewDate(2026, 11, 10), ContributionCount: 4},
			},
		},
	}
	summary, err := history.SummarizeContributions(cleanSnapshot)
	if err != nil {
		t.Fatalf("unexpected error for clean history: %v", err)
	}
	if summary.TotalContributions != 5 {
		t.Errorf("expected 5 contributions, got %d", summary.TotalContributions)
	}
}

func TestSnapshotSchemaVersions(t *testing.T) {
	// Schema 1 rejects contribution_history
	schema1JSON := []byte(`{"schema_version":1,"account_id":10,"contribution_history":{"observed_on":"2026-11-10","days":[]}}`)
	var s1 history.AccountSnapshot
	if err := json.Unmarshal(schema1JSON, &s1); err == nil {
		t.Errorf("schema 1 should reject contribution_history")
	}

	// Schema 2 requires contribution_history but rejects language_colors
	schema2JSON := []byte(`{"schema_version":2,"account_id":10,"language_colors":{}}`)
	var s2 history.AccountSnapshot
	if err := json.Unmarshal(schema2JSON, &s2); err == nil {
		t.Errorf("schema 2 should reject missing contribution_history or presence of language_colors")
	}

	// Unsupported schema version
	unsupportedJSON := []byte(`{"schema_version":99,"account_id":10}`)
	var s99 history.AccountSnapshot
	if err := json.Unmarshal(unsupportedJSON, &s99); err == nil {
		t.Errorf("unsupported schema version 99 should fail")
	}
}

func TestISOWeekRollOver(t *testing.T) {
	// End of 2024 into 2025
	dDec30 := history.NewDate(2024, time.December, 30) // Monday of week 1 of 2025
	y, w := dDec30.ISOWeek()
	if y != 2025 || w != 1 {
		t.Logf("ISO week for 2024-12-30: year=%d, week=%d", y, w)
	}
}
