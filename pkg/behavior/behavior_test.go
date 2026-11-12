package behavior_test

import (
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/behavior"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
)

func TestBehaviorStrings(t *testing.T) {
	cases := map[behavior.Behavior]string{
		behavior.DeepRest:       "deep rest",
		behavior.Rest:           "rest",
		behavior.SurfaceFeeding: "surface feeding",
		behavior.Dart:           "dart",
		behavior.Cruise:         "cruise",
		behavior.Behavior(99):   "unknown",
	}

	for b, expected := range cases {
		if b.String() != expected {
			t.Errorf("expected %s, got %s", expected, b.String())
		}
	}
}

func TestSelectBehaviorQuietDays(t *testing.T) {
	// 1. Deep rest (>= 60 quiet days)
	devDeepRest := &koi.KoiDevelopment{
		RecentActivity: koi.RecentActivity{
			QuietDayCount: 65,
		},
	}
	if b := behavior.SelectBehavior(devDeepRest); b != behavior.DeepRest {
		t.Errorf("expected DeepRest, got %s", b.String())
	}

	// 2. Rest (>= 14 quiet days)
	devRest := &koi.KoiDevelopment{
		RecentActivity: koi.RecentActivity{
			QuietDayCount: 20,
		},
	}
	if b := behavior.SelectBehavior(devRest); b != behavior.Rest {
		t.Errorf("expected Rest, got %s", b.String())
	}
}

func TestSelectBehaviorSurfaceFeeding(t *testing.T) {
	// First contribution of week (previous days of week 0, current day > 0)
	devFeeding := &koi.KoiDevelopment{
		RecentActivity: koi.RecentActivity{
			QuietDayCount:                0,
			ObservationWeekContributions: []uint64{0, 0, 5},
		},
	}
	if b := behavior.SelectBehavior(devFeeding); b != behavior.SurfaceFeeding {
		t.Errorf("expected SurfaceFeeding, got %s", b.String())
	}
}

func TestSelectBehaviorDartMilestone(t *testing.T) {
	// Crossed milestone 1000
	devDart := &koi.KoiDevelopment{
		TotalContributions: 1005,
		RecentActivity: koi.RecentActivity{
			QuietDayCount:                0,
			RecentContributions:          20,
			ObservationWeekContributions: []uint64{2, 3, 5},
		},
	}
	if b := behavior.SelectBehavior(devDart); b != behavior.Dart {
		t.Errorf("expected Dart when crossing milestone 1000, got %s", b.String())
	}
}

func TestSelectBehaviorCruiseDefault(t *testing.T) {
	devCruise := &koi.KoiDevelopment{
		TotalContributions: 450,
		RecentActivity: koi.RecentActivity{
			QuietDayCount:                0,
			RecentContributions:          10,
			ObservationWeekContributions: []uint64{1, 2, 3},
		},
	}
	if b := behavior.SelectBehavior(devCruise); b != behavior.Cruise {
		t.Errorf("expected Cruise, got %s", b.String())
	}
	_ = history.DateFormat
}
