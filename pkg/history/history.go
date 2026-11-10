package history

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// LanguageColor represents a 24-bit RGB color.
type LanguageColor [3]byte

// ParseLanguageColor parses a hex color string in #rrggbb format.
func ParseLanguageColor(s string) (LanguageColor, error) {
	if !strings.HasPrefix(s, "#") || len(s) != 7 {
		return LanguageColor{}, fmt.Errorf("language color %s is not #rrggbb", s)
	}
	bytes, err := hex.DecodeString(s[1:])
	if err != nil {
		return LanguageColor{}, fmt.Errorf("language color %s is not #rrggbb", s)
	}
	return LanguageColor{bytes[0], bytes[1], bytes[2]}, nil
}

func (c LanguageColor) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

func (c LanguageColor) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Hex())
}

func (c *LanguageColor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseLanguageColor(s)
	if err != nil {
		return err
	}
	*c = parsed
	return nil
}

// RepositoryVisibility defines repository visibility as reported by GitHub API.
type RepositoryVisibility string

const (
	VisibilityPublic   RepositoryVisibility = "public"
	VisibilityPrivate  RepositoryVisibility = "private"
	VisibilityInternal RepositoryVisibility = "internal"
)

// RepositoryObservation holds observed repository metadata.
type RepositoryObservation struct {
	RepositoryID    uint64               `json:"repository_id"`
	Visibility      RepositoryVisibility `json:"visibility"`
	IsFork          bool                 `json:"is_fork"`
	PrimaryLanguage *string              `json:"primary_language"`
}

// Date holds a UTC calendar date (YYYY-MM-DD).
type Date struct {
	time.Time
}

const DateFormat = "2006-01-02"

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

func ParseDate(s string) (Date, error) {
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return Date{}, err
	}
	return Date{t.UTC()}, nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(DateFormat))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

func (d Date) String() string {
	return d.Format(DateFormat)
}

func (d Date) AddDays(days int) Date {
	return Date{d.Time.AddDate(0, 0, days)}
}

// ContributionDay represents contributions on a single day.
type ContributionDay struct {
	Date              Date   `json:"date"`
	ContributionCount uint32 `json:"contribution_count"`
}

// ContributionHistory holds daily contributions through an observation date.
type ContributionHistory struct {
	ObservedOn Date              `json:"observed_on"`
	Days       []ContributionDay `json:"days"`
}

// ISOWeekKey represents an ISO 8601 year and week.
type ISOWeekKey struct {
	Year int
	Week int
}

// ContributionWeek represents participation in one Monday-Sunday week.
type ContributionWeek struct {
	Year             int
	Week             int
	ActiveDayCount   uint8
	ObservedDayCount uint8
}

// ContributionSummary holds aggregated contribution statistics.
type ContributionSummary struct {
	ObservedOn         Date
	TotalContributions uint64
	Weeks              []ContributionWeek
	DailyCounts        map[string]uint32
}

// AccountSnapshot holds GitHub account observations.
type AccountSnapshot struct {
	AccountID           uint64                             `json:"account_id"`
	Repositories        []RepositoryObservation            `json:"repositories"`
	ContributionHistory *ContributionHistory               `json:"contribution_history,omitempty"`
	LanguageColors      map[string]*LanguageColor          `json:"language_colors,omitempty"`
	SchemaVersion       uint32                             `json:"schema_version"`
}

type snapshotDocument struct {
	SchemaVersion       uint32                    `json:"schema_version"`
	AccountID           uint64                    `json:"account_id"`
	Repositories        []RepositoryObservation   `json:"repositories"`
	ContributionHistory *ContributionHistory      `json:"contribution_history,omitempty"`
	LanguageColors      map[string]*LanguageColor `json:"language_colors,omitempty"`
}

func (s *AccountSnapshot) UnmarshalJSON(data []byte) error {
	var doc snapshotDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}

	hasHistory := doc.ContributionHistory != nil
	hasColors := doc.LanguageColors != nil

	switch doc.SchemaVersion {
	case 1:
		if hasHistory {
			return errors.New("contribution_history is not a field in snapshot version 1; use version 2 for dated contributions")
		}
		if hasColors {
			return errors.New("language_colors is not a field before snapshot version 3; use version 3 for language colors")
		}
	case 2:
		if !hasHistory {
			return errors.New("missing contribution_history in snapshot versions 2 and 3; supply dated observations")
		}
		if hasColors {
			return errors.New("language_colors is not a field before snapshot version 3; use version 3 for language colors")
		}
	case 3:
		if !hasHistory {
			return errors.New("missing contribution_history in snapshot versions 2 and 3; supply dated observations")
		}
		if !hasColors {
			return errors.New("missing language_colors in snapshot version 3; collect the snapshot again")
		}
	default:
		return fmt.Errorf("unsupported snapshot schema version %d; use a reader supporting that version", doc.SchemaVersion)
	}

	s.AccountID = doc.AccountID
	s.Repositories = doc.Repositories
	s.ContributionHistory = doc.ContributionHistory
	s.LanguageColors = doc.LanguageColors
	s.SchemaVersion = doc.SchemaVersion
	return nil
}

func (s AccountSnapshot) MarshalJSON() ([]byte, error) {
	ver := s.SchemaVersion
	if ver == 0 {
		if s.LanguageColors != nil {
			ver = 3
		} else if s.ContributionHistory != nil {
			ver = 2
		} else {
			ver = 1
		}
	}
	doc := snapshotDocument{
		SchemaVersion:       ver,
		AccountID:           s.AccountID,
		Repositories:        s.Repositories,
		ContributionHistory: s.ContributionHistory,
		LanguageColors:      s.LanguageColors,
	}
	return json.Marshal(doc)
}

// CalculateLanguageShares computes primary language shares across public, non-fork repositories.
func CalculateLanguageShares(snapshot *AccountSnapshot) (map[string]float64, error) {
	distinct := make(map[uint64]RepositoryObservation)
	for _, repo := range snapshot.Repositories {
		if prev, exists := distinct[repo.RepositoryID]; exists {
			if prev.Visibility != repo.Visibility || prev.IsFork != repo.IsFork {
				return nil, fmt.Errorf("conflicting observations for repository %d; collect a consistent snapshot", repo.RepositoryID)
			}
			if (prev.PrimaryLanguage == nil) != (repo.PrimaryLanguage == nil) {
				return nil, fmt.Errorf("conflicting observations for repository %d; collect a consistent snapshot", repo.RepositoryID)
			}
			if prev.PrimaryLanguage != nil && *prev.PrimaryLanguage != *repo.PrimaryLanguage {
				return nil, fmt.Errorf("conflicting observations for repository %d; collect a consistent snapshot", repo.RepositoryID)
			}
		} else {
			distinct[repo.RepositoryID] = repo
		}
	}

	langCounts := make(map[string]int)
	eligibleTotal := 0
	for _, repo := range distinct {
		if repo.Visibility != VisibilityPublic || repo.IsFork {
			continue
		}
		if repo.PrimaryLanguage != nil && *repo.PrimaryLanguage != "" {
			langCounts[*repo.PrimaryLanguage]++
			eligibleTotal++
		}
	}

	shares := make(map[string]float64)
	if eligibleTotal == 0 {
		return shares, nil
	}
	for lang, count := range langCounts {
		shares[lang] = float64(count) / float64(eligibleTotal)
	}
	return shares, nil
}

// SummarizeContributions validates and groups daily contributions into chronological ISO weeks.
func SummarizeContributions(snapshot *AccountSnapshot) (*ContributionSummary, error) {
	if snapshot.ContributionHistory == nil {
		return nil, errors.New("contribution history is unavailable; supply a snapshot with dated contributions")
	}
	history := snapshot.ContributionHistory

	dailyCounts := make(map[string]uint32)
	var dates []Date

	for _, day := range history.Days {
		if day.Date.After(history.ObservedOn.Time) {
			return nil, fmt.Errorf("contribution date %s is after observation date %s; supply observations through the stated date", day.Date, history.ObservedOn)
		}
		dateStr := day.Date.String()
		if prevCount, exists := dailyCounts[dateStr]; exists {
			if prevCount != day.ContributionCount {
				return nil, fmt.Errorf("conflicting contribution counts for %s; collect a consistent history", day.Date)
			}
		} else {
			dailyCounts[dateStr] = day.ContributionCount
			dates = append(dates, day.Date)
		}
	}

	if len(dates) == 0 {
		return &ContributionSummary{
			ObservedOn:         history.ObservedOn,
			TotalContributions: 0,
			Weeks:              nil,
			DailyCounts:        dailyCounts,
		}, nil
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j].Time)
	})

	// Check continuous days without missing dates
	for i := 0; i < len(dates)-1; i++ {
		expected := dates[i].AddDays(1)
		if !expected.Equal(dates[i+1].Time) {
			return nil, fmt.Errorf("missing contribution date %s; include observations for every date, including zero-activity days", expected)
		}
	}
	// Also ensure history reaches observed_on
	lastDate := dates[len(dates)-1]
	if lastDate.Before(history.ObservedOn.Time) {
		expected := lastDate.AddDays(1)
		return nil, fmt.Errorf("missing contribution date %s; include observations for every date, including zero-activity days", expected)
	}

	type weekKey struct {
		year int
		week int
	}
	var weekOrder []weekKey
	weeklyMap := make(map[weekKey]*ContributionWeek)

	var total uint64
	for _, d := range dates {
		count := dailyCounts[d.String()]
		total += uint64(count)

		year, week := d.ISOWeek()
		k := weekKey{year, week}
		cw, exists := weeklyMap[k]
		if !exists {
			cw = &ContributionWeek{
				Year: year,
				Week: week,
			}
			weeklyMap[k] = cw
			weekOrder = append(weekOrder, k)
		}
		if count > 0 {
			cw.ActiveDayCount++
		}
		cw.ObservedDayCount++
	}

	weeks := make([]ContributionWeek, len(weekOrder))
	for i, k := range weekOrder {
		weeks[i] = *weeklyMap[k]
	}

	return &ContributionSummary{
		ObservedOn:         history.ObservedOn,
		TotalContributions: total,
		Weeks:              weeks,
		DailyCounts:        dailyCounts,
	}, nil
}
