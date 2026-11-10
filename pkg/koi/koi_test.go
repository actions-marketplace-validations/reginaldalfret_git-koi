package koi_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reginaldalfret/git-koi/pkg/history"
	"github.com/reginaldalfret/git-koi/pkg/koi"
)

func TestDeriveKoiDevelopment(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "account-with-history.json")
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var snapshot history.AccountSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		t.Fatalf("failed to unmarshal snapshot: %v", err)
	}

	dev, err := koi.DeriveKoiDevelopment(&snapshot)
	if err != nil {
		t.Fatalf("failed to derive koi development: %v", err)
	}

	if dev.IdentitySeed != 1001 {
		t.Errorf("expected seed 1001, got %d", dev.IdentitySeed)
	}

	if dev.TotalContributions != 14 {
		t.Errorf("expected 14 contributions, got %d", dev.TotalContributions)
	}

	growth := koi.CalculateGrowthFraction(dev.TotalContributions)
	if growth <= 0 || growth >= 1.0 {
		t.Errorf("expected growth in (0, 1), got %f", growth)
	}
}
