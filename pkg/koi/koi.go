package koi

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
)

const (
	FullColorContributions  uint64 = 500
	YoungShapeContributions uint64 = 1500
	AdultContributions      uint64 = 12750
	MaxGrowthContributions  uint64 = 100000
)

// CalculateGrowthFraction maps total contributions to a growth fraction [0.0, 1.0].
func CalculateGrowthFraction(totalContributions uint64) float64 {
	const contributionScale = 50.0
	bounded := float64(totalContributions)
	if totalContributions > MaxGrowthContributions {
		bounded = float64(MaxGrowthContributions)
	}
	return math.Log2(1.0+bounded/contributionScale) / math.Log2(1.0+float64(MaxGrowthContributions)/contributionScale)
}

// RecentActivity records recent contribution frequency around the observation date.
type RecentActivity struct {
	QuietDayCount                 uint32
	RecentContributions           uint64
	CompleteWeekContributions     []uint64
	ObservationWeekContributions  []uint64
}

// RecentStreakType denotes whether the most recent streak is current or ended.
type RecentStreakType int

const (
	StreakCurrent RecentStreakType = iota
	StreakEnded
)

// RecentStreak represents the latest weekly streak.
type RecentStreak struct {
	Type    RecentStreakType
	Weeks   []history.ContributionWeek
	EndedOn history.Date
}

// StreakDevelopment holds longest and most recent weekly streaks.
type StreakDevelopment struct {
	Longest []history.ContributionWeek
	Recent  *RecentStreak
}

// KoiDevelopment represents the derived biological identity and visual traits of the koi.
type KoiDevelopment struct {
	IdentitySeed       uint64
	LanguageShares     map[string]float64
	LanguageColors     map[string]*history.LanguageColor
	TotalContributions uint64
	ObservedOn         history.Date
	Streaks            StreakDevelopment
	RecentActivity     RecentActivity
}

// DeriveKoiDevelopment calculates reproducible koi development from account observations.
func DeriveKoiDevelopment(snapshot *history.AccountSnapshot) (*KoiDevelopment, error) {
	summary, err := history.SummarizeContributions(snapshot)
	if err != nil {
		return nil, err
	}

	shares, err := history.CalculateLanguageShares(snapshot)
	if err != nil {
		return nil, err
	}

	if snapshot.LanguageColors == nil {
		return nil, errors.New("the snapshot has no language colors; collect it again with `git-koi collect`")
	}

	colors := make(map[string]*history.LanguageColor)
	for lang := range shares {
		c, ok := snapshot.LanguageColors[lang]
		if !ok {
			return nil, fmt.Errorf("the snapshot has no color for %s; collect it again with `git-koi collect`", lang)
		}
		colors[lang] = c
	}

	return &KoiDevelopment{
		IdentitySeed:       snapshot.AccountID,
		LanguageShares:     shares,
		LanguageColors:     colors,
		TotalContributions: summary.TotalContributions,
		ObservedOn:         summary.ObservedOn,
		Streaks:            calculateStreakDevelopment(summary),
		RecentActivity:     measureRecentActivity(summary),
	}, nil
}

func measureRecentActivity(summary *history.ContributionSummary) RecentActivity {
	observedOn := summary.ObservedOn
	dailyCounts := summary.DailyCounts

	// Sort dates descending to count quiet days
	var dateKeys []string
	for d := range dailyCounts {
		dateKeys = append(dateKeys, d)
	}
	sort.Strings(dateKeys)

	quietCount := uint32(0)
	for i := len(dateKeys) - 1; i >= 0; i-- {
		if dailyCounts[dateKeys[i]] == 0 {
			quietCount++
		} else {
			break
		}
	}

	// Recent 7 days contributions ending on observed_on
	recentStart := observedOn.AddDays(-6)
	var recentContributions uint64
	for d := recentStart; !d.After(observedOn.Time); d = d.AddDays(1) {
		recentContributions += uint64(dailyCounts[d.String()])
	}

	// Group counts by ISO week
	type weekInfo struct {
		contributions uint64
		dayCount      int
	}
	weeklyTotals := make(map[history.ISOWeekKey]*weekInfo)
	var weekOrder []history.ISOWeekKey

	for _, k := range dateKeys {
		d, _ := history.ParseDate(k)
		y, w := d.ISOWeek()
		key := history.ISOWeekKey{Year: y, Week: w}
		info, exists := weeklyTotals[key]
		if !exists {
			info = &weekInfo{}
			weeklyTotals[key] = info
			weekOrder = append(weekOrder, key)
		}
		info.contributions += uint64(dailyCounts[k])
		info.dayCount++
	}

	obsYear, obsWeek := observedOn.ISOWeek()
	obsWeekKey := history.ISOWeekKey{Year: obsYear, Week: obsWeek}

	var completeWeeks []uint64
	for _, k := range weekOrder {
		info := weeklyTotals[k]
		if info.dayCount == 7 && k != obsWeekKey {
			completeWeeks = append(completeWeeks, info.contributions)
		}
	}

	// Days of observation week from Monday through observed_on
	weekdayOffset := int(observedOn.Weekday())
	if weekdayOffset == 0 {
		weekdayOffset = 7 // Sunday is day 7 from Monday
	}
	obsMonday := observedOn.AddDays(-(weekdayOffset - 1))

	var obsWeekContributions []uint64
	for d := obsMonday; !d.After(observedOn.Time); d = d.AddDays(1) {
		obsWeekContributions = append(obsWeekContributions, uint64(dailyCounts[d.String()]))
	}

	return RecentActivity{
		QuietDayCount:                quietCount,
		RecentContributions:          recentContributions,
		CompleteWeekContributions:    completeWeeks,
		ObservationWeekContributions: obsWeekContributions,
	}
}

func calculateStreakDevelopment(summary *history.ContributionSummary) StreakDevelopment {
	longestStart, longestEnd := 0, 0
	recentStart, recentEnd := 0, 0
	var currentStart *int
	var endedOn *history.Date

	obsYear, obsWeek := summary.ObservedOn.ISOWeek()

	for weekIdx, cw := range summary.Weeks {
		if cw.ActiveDayCount > 0 {
			if currentStart == nil {
				val := weekIdx
				currentStart = &val
			}
			recentStart = *currentStart
			recentEnd = weekIdx + 1
			endedOn = nil

			if (recentEnd - recentStart) > (longestEnd - longestStart) {
				longestStart = recentStart
				longestEnd = recentEnd
			}
		} else if (cw.Year != obsYear || cw.Week != obsWeek) && currentStart != nil {
			currentStart = nil
			// The Sunday of that inactive ISO week
			inactiveSunday := isoWeekSunday(cw.Year, cw.Week)
			nextMonday := inactiveSunday.AddDays(1)
			endedOn = &nextMonday
		}
	}

	var recentStreak *RecentStreak
	if recentEnd > recentStart {
		weeks := make([]history.ContributionWeek, recentEnd-recentStart)
		copy(weeks, summary.Weeks[recentStart:recentEnd])
		if endedOn != nil {
			recentStreak = &RecentStreak{
				Type:    StreakEnded,
				Weeks:   weeks,
				EndedOn: *endedOn,
			}
		} else {
			recentStreak = &RecentStreak{
				Type:  StreakCurrent,
				Weeks: weeks,
			}
		}
	}

	var longestWeeks []history.ContributionWeek
	if longestEnd > longestStart {
		longestWeeks = make([]history.ContributionWeek, longestEnd-longestStart)
		copy(longestWeeks, summary.Weeks[longestStart:longestEnd])
	}

	return StreakDevelopment{
		Longest: longestWeeks,
		Recent:  recentStreak,
	}
}

// isoWeekSunday calculates the calendar date of the Sunday of an ISO week.
func isoWeekSunday(year, week int) history.Date {
	// Jan 4th is always in ISO week 1
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	weekdayOffset := int(jan4.Weekday())
	if weekdayOffset == 0 {
		weekdayOffset = 7
	}
	// Monday of week 1
	week1Monday := jan4.AddDate(0, 0, -(weekdayOffset - 1))
	// Target week Monday
	targetMonday := week1Monday.AddDate(0, 0, (week-1)*7)
	// Sunday is 6 days after Monday
	return history.Date{Time: targetMonday.AddDate(0, 0, 6)}
}
