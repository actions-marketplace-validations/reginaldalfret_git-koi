package behavior

import (
	"github.com/reginaldalfret/git-koi/pkg/koi"
)

// Behavior defines the day's movement style.
type Behavior int

const (
	DeepRest Behavior = iota
	Rest
	SurfaceFeeding
	Dart
	Cruise
)

func (b Behavior) String() string {
	switch b {
	case DeepRest:
		return "deep rest"
	case Rest:
		return "rest"
	case SurfaceFeeding:
		return "surface feeding"
	case Dart:
		return "dart"
	case Cruise:
		return "cruise"
	default:
		return "unknown"
	}
}

const (
	DeepRestQuietDays        uint32 = 60
	RestQuietDays            uint32 = 14
	DartComparedWeeks        int    = 52
	DartMinCompleteWeeks     int    = 10
	DartWeekShareDenominator int    = 10
	DaysPerWeek              int    = 7
)

var DartMilestones = []uint64{1000, 5000, 10000, 25000, 50000, 100000}

// SelectBehavior chooses the movement for the day after the observation date.
func SelectBehavior(dev *koi.KoiDevelopment) Behavior {
	activity := &dev.RecentActivity
	if activity.QuietDayCount >= DeepRestQuietDays {
		return DeepRest
	} else if activity.QuietDayCount >= RestQuietDays {
		return Rest
	} else if isFirstContributionOfWeek(activity) {
		return SurfaceFeeding
	} else if hasCrossedMilestone(dev.TotalContributions, activity) || isFinishedBusyWeek(activity) {
		return Dart
	}
	return Cruise
}

func isFirstContributionOfWeek(activity *koi.RecentActivity) bool {
	contribs := activity.ObservationWeekContributions
	if len(contribs) == 0 {
		return false
	}
	lastIdx := len(contribs) - 1
	if contribs[lastIdx] == 0 {
		return false
	}
	for i := 0; i < lastIdx; i++ {
		if contribs[i] > 0 {
			return false
		}
	}
	return true
}

func hasCrossedMilestone(total uint64, activity *koi.RecentActivity) bool {
	var earlierTotal uint64
	if total > activity.RecentContributions {
		earlierTotal = total - activity.RecentContributions
	}
	for _, m := range DartMilestones {
		if earlierTotal < m && m <= total {
			return true
		}
	}
	return false
}

func isFinishedBusyWeek(activity *koi.RecentActivity) bool {
	finishedWeek := activity.ObservationWeekContributions
	completeWeeks := activity.CompleteWeekContributions
	if len(finishedWeek) != DaysPerWeek {
		return false
	}
	start := 0
	if len(completeWeeks) > DartComparedWeeks {
		start = len(completeWeeks) - DartComparedWeeks
	}
	compared := completeWeeks[start:]
	if len(compared) < DartMinCompleteWeeks {
		return false
	}
	var weekTotal uint64
	for _, c := range finishedWeek {
		weekTotal += c
	}
	matchCount := 0
	for _, wc := range compared {
		if wc >= weekTotal {
			matchCount++
		}
	}
	return matchCount*DartWeekShareDenominator <= len(compared)
}
