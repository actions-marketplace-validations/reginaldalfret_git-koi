package markings_test

import (
	"testing"
	"time"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/koi"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/markings"
)

func TestCalculateStreakMarks(t *testing.T) {
	obs := history.NewDate(2026, 11, 10)

	// 1. No streaks
	devNoStreak := &koi.KoiDevelopment{
		ObservedOn: obs,
		Streaks:    koi.StreakDevelopment{},
	}
	marks := markings.CalculateStreakMarks(devNoStreak)
	if marks.StreakFinStrength != 0.0 {
		t.Errorf("expected 0.0 streak fin strength without streak, got %f", marks.StreakFinStrength)
	}

	// 2. Active current streak (15 weeks > 13 full weeks)
	weeks15 := make([]history.ContributionWeek, 15)
	devActiveStreak := &koi.KoiDevelopment{
		ObservedOn: obs,
		Streaks: koi.StreakDevelopment{
			Recent: &koi.RecentStreak{
				Type:  koi.StreakCurrent,
				Weeks: weeks15,
			},
			Longest: weeks15,
		},
	}
	marksActive := markings.CalculateStreakMarks(devActiveStreak)
	if marksActive.StreakFinStrength != 1.0 {
		t.Errorf("expected 1.0 fin strength for >=13 week streak, got %f", marksActive.StreakFinStrength)
	}
	if marksActive.ColorVibrancy <= 1.0 {
		t.Errorf("expected vibrancy > 1.0 for 15 week longest streak, got %f", marksActive.ColorVibrancy)
	}

	// 3. Ended streak with fade (ended 2 weeks ago)
	endedOn := history.Date{Time: obs.Time.AddDate(0, 0, -14)}
	devEndedStreak := &koi.KoiDevelopment{
		ObservedOn: obs,
		Streaks: koi.StreakDevelopment{
			Recent: &koi.RecentStreak{
				Type:    koi.StreakEnded,
				Weeks:   weeks15,
				EndedOn: endedOn,
			},
			Longest: weeks15,
		},
	}
	marksEnded := markings.CalculateStreakMarks(devEndedStreak)
	if marksEnded.StreakFinStrength >= 1.0 || marksEnded.StreakFinStrength <= 0.0 {
		t.Errorf("expected faded fin strength in (0, 1), got %f", marksEnded.StreakFinStrength)
	}

	// 4. Legendary long streak (> 208 weeks for full glimmer)
	weeks300 := make([]history.ContributionWeek, 300)
	devLegendary := &koi.KoiDevelopment{
		ObservedOn: obs,
		Streaks: koi.StreakDevelopment{
			Longest: weeks300,
		},
	}
	marksLegendary := markings.CalculateStreakMarks(devLegendary)
	if marksLegendary.GlimmerStrength != 1.0 {
		t.Errorf("expected max glimmer 1.0 for 300 week streak, got %f", marksLegendary.GlimmerStrength)
	}
	_ = time.RFC3339
}
