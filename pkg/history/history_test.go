package history_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
