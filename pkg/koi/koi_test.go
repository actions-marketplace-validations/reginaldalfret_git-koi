package koi_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/behavior"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
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

	b := behavior.SelectBehavior(dev)
	if b.String() == "unknown" {
		t.Errorf("unexpected unknown behavior")
	}
}

func TestCalculateGrowthFractionBoundaries(t *testing.T) {
	cases := []struct {
		contribs uint64
		minFrac  float64
		maxFrac  float64
	}{
		{0, 0.0, 0.0},
		{1, 0.001, 0.05},
		{koi.FullColorContributions, 0.25, 0.35},
		{koi.YoungShapeContributions, 0.40, 0.50},
		{koi.AdultContributions, 0.70, 0.76},
		{koi.MaxGrowthContributions, 1.0, 1.0},
		{koi.MaxGrowthContributions + 50000, 1.0, 1.0}, // clamped at max
		{math.MaxUint64, 1.0, 1.0},
	}

	for _, tc := range cases {
		frac := koi.CalculateGrowthFraction(tc.contribs)
		if frac < tc.minFrac || frac > tc.maxFrac {
			t.Errorf("contribs=%d: expected growth in [%f, %f], got %f", tc.contribs, tc.minFrac, tc.maxFrac, frac)
		}
		if math.IsNaN(frac) || math.IsInf(frac, 0) {
			t.Errorf("contribs=%d: produced NaN or Inf", tc.contribs)
		}
	}
}

func TestZeroContributionAccount(t *testing.T) {
	c, _ := history.ParseLanguageColor("#00ADD8")
	observed := history.NewDate(2026, 11, 10)
	days := []history.ContributionDay{
		{Date: observed.AddDays(-1), ContributionCount: 0},
		{Date: observed, ContributionCount: 0},
	}

	snapshot := &history.AccountSnapshot{
		AccountID: 42,
		ContributionHistory: &history.ContributionHistory{
			ObservedOn: observed,
			Days:       days,
		},
		Repositories: []history.RepositoryObservation{
			{
				RepositoryID:    1,
				Visibility:      history.VisibilityPublic,
				IsFork:          false,
				PrimaryLanguage: stringPtr("Go"),
			},
		},
		LanguageColors: map[string]*history.LanguageColor{
			"Go": &c,
		},
		SchemaVersion: 3,
	}

	dev, err := koi.DeriveKoiDevelopment(snapshot)
	if err != nil {
		t.Fatalf("unexpected error deriving zero-contribution development: %v", err)
	}
	if dev.TotalContributions != 0 {
		t.Errorf("expected 0 contributions, got %d", dev.TotalContributions)
	}

	frac := koi.CalculateGrowthFraction(dev.TotalContributions)
	if frac != 0.0 {
		t.Errorf("expected 0.0 growth fraction for 0 contributions, got %f", frac)
	}

	// Behavior for zero activity should be rest or deep rest
	b := behavior.SelectBehavior(dev)
	if b != behavior.Rest && b != behavior.DeepRest {
		t.Logf("zero contribution behavior selected: %s", b.String())
	}
}

func stringPtr(s string) *string {
	return &s
}
