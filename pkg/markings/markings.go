package markings

import (
	"math"

	"github.com/reginaldalfret/git-koi/pkg/koi"
)

const (
	StreakFinFullWeeks = 13.0
	FadeRatePerWeek    = 0.5
	FadeWeeks          = 6.5
	RecordCurveWeeks   = 0.6
	NormalColorWeeks   = 4.0
	VividColorWeeks    = 52.0
	FullGlimmerWeeks   = 208.0
)

// StreakMarks holds visual mark strengths for the renderer.
type StreakMarks struct {
	StreakFinStrength float64
	ColorVibrancy     float64
	GlimmerStrength   float64
}

// CalculateStreakMarks derives streak mark strengths from koi development.
func CalculateStreakMarks(dev *koi.KoiDevelopment) StreakMarks {
	var streakFinStrength float64
	if dev.Streaks.Recent == nil {
		streakFinStrength = 0.0
	} else if dev.Streaks.Recent.Type == koi.StreakCurrent {
		streakFinStrength = calculateFinBuild(len(dev.Streaks.Recent.Weeks))
	} else {
		// StreakEnded
		endedOn := dev.Streaks.Recent.EndedOn
		days := dev.ObservedOn.Sub(endedOn.Time).Hours() / 24.0
		if days < 0 {
			days = 0
		}
		weeksSinceEnd := days / 7.0
		streakFinStrength = calculateFinBuild(len(dev.Streaks.Recent.Weeks)) * calculateFade(weeksSinceEnd)
	}

	recordStrength := calculateRecordStrength(float64(len(dev.Streaks.Longest)))
	normalStrength := calculateRecordStrength(NormalColorWeeks)
	vividStrength := calculateRecordStrength(VividColorWeeks)

	var colorVibrancy float64
	if recordStrength < normalStrength {
		colorVibrancy = recordStrength / normalStrength
	} else {
		extra := (recordStrength - normalStrength) / (vividStrength - normalStrength)
		if extra > 1.0 {
			extra = 1.0
		}
		colorVibrancy = 1.0 + extra
	}

	glimmer := (recordStrength - vividStrength) / (1.0 - vividStrength)
	if glimmer < 0.0 {
		glimmer = 0.0
	} else if glimmer > 1.0 {
		glimmer = 1.0
	}

	return StreakMarks{
		StreakFinStrength: streakFinStrength,
		ColorVibrancy:     colorVibrancy,
		GlimmerStrength:   glimmer,
	}
}

func calculateFinBuild(streakWeekCount int) float64 {
	val := float64(streakWeekCount) / StreakFinFullWeeks
	if val > 1.0 {
		return 1.0
	}
	return val
}

func calculateFade(weeksSinceEnd float64) float64 {
	fadeFloor := math.Exp(-FadeRatePerWeek * FadeWeeks)
	val := (math.Exp(-FadeRatePerWeek*weeksSinceEnd) - fadeFloor) / (1.0 - fadeFloor)
	if val < 0.0 {
		return 0.0
	}
	return val
}

func calculateRecordStrength(longestWeeks float64) float64 {
	val := math.Log(1.0+longestWeeks/RecordCurveWeeks) / math.Log(1.0+FullGlimmerWeeks/RecordCurveWeeks)
	if val > 1.0 {
		return 1.0
	}
	return val
}
