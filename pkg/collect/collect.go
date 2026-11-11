package collect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
)

// GraphqlClient runs a GraphQL query against GitHub's API.
type GraphqlClient interface {
	RunQuery(query string, variables map[string]string) (map[string]any, error)
}

// GhCli executes GraphQL queries using the GitHub CLI (`gh api graphql`).
type GhCli struct{}

func (g GhCli) RunQuery(query string, variables map[string]string) (map[string]any, error) {
	args := []string{"api", "graphql", "-f", fmt.Sprintf("query=%s", query)}
	for k, v := range variables {
		args = append(args, "-f", fmt.Sprintf("%s=%s", k, v))
	}

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if strings.Contains(err.Error(), "executable file not found") {
			return nil, errors.New("GitHub CLI (gh) is not installed; install it from https://cli.github.com")
		}
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return nil, fmt.Errorf("gh command failed: %s", errMsg)
		}
		return nil, fmt.Errorf("gh command failed: %w", err)
	}

	var root map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &root); err != nil {
		return nil, fmt.Errorf("gh output is not valid JSON: %w", err)
	}

	data, ok := root["data"].(map[string]any)
	if !ok {
		if errs, hasErrs := root["errors"]; hasErrs {
			return nil, fmt.Errorf("graphql error: %v", errs)
		}
		return nil, errors.New("response is missing data object")
	}
	return data, nil
}

const (
	accountQuery = `query($login: String!) { user(login: $login) { databaseId createdAt } }`

	repositoryFields = `nodes { databaseId visibility isFork primaryLanguage { name color } } pageInfo { hasNextPage endCursor }`

	calendarQuery = `query($login: String!, $from: DateTime!, $to: DateTime!) {
		user(login: $login) { contributionsCollection(from: $from, to: $to) {
		contributionCalendar { weeks { contributionDays { date contributionCount } } } } } }`
)

// CollectAccountSnapshot gathers public repositories, language colors, and complete contribution history.
func CollectAccountSnapshot(client GraphqlClient, login string, today history.Date) (*history.AccountSnapshot, error) {
	data, err := client.RunQuery(accountQuery, map[string]string{"login": login})
	if err != nil {
		return nil, err
	}

	user, ok := data["user"].(map[string]any)
	if !ok || user == nil {
		return nil, fmt.Errorf("account %s was not found", login)
	}

	accountIDFloat, ok := user["databaseId"].(float64)
	if !ok {
		return nil, errors.New("databaseId is missing or not a number")
	}
	accountID := uint64(accountIDFloat)

	createdAtStr, ok := user["createdAt"].(string)
	if !ok {
		return nil, errors.New("createdAt is missing or not a string")
	}
	createdAtTime, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("createdAt is not a valid timestamp: %w", err)
	}
	createdOn := history.NewDate(createdAtTime.Year(), createdAtTime.Month(), createdAtTime.Day())

	languageColors := make(map[string]*history.LanguageColor)

	ownedRepos, err := collectRepositories(client, login, languageColors, "repositories",
		"repositories(first: 100, after: $cursor, ownerAffiliations: [OWNER], privacy: PUBLIC)")
	if err != nil {
		return nil, err
	}

	contribRepos, err := collectRepositories(client, login, languageColors, "repositoriesContributedTo",
		"repositoriesContributedTo(first: 100, after: $cursor, privacy: PUBLIC, includeUserRepositories: false, contributionTypes: [COMMIT, ISSUE, PULL_REQUEST, PULL_REQUEST_REVIEW, REPOSITORY])")
	if err != nil {
		return nil, err
	}

	allRepos := append(ownedRepos, contribRepos...)

	historyObj, err := collectContributionHistory(client, login, createdOn, today)
	if err != nil {
		return nil, err
	}

	snapshot := &history.AccountSnapshot{
		AccountID:           accountID,
		Repositories:        allRepos,
		ContributionHistory: historyObj,
		LanguageColors:      languageColors,
		SchemaVersion:       3,
	}

	if _, err := history.SummarizeContributions(snapshot); err != nil {
		return nil, err
	}
	if _, err := history.CalculateLanguageShares(snapshot); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func collectRepositories(client GraphqlClient, login string, languageColors map[string]*history.LanguageColor, connName, connCall string) ([]history.RepositoryObservation, error) {
	query := fmt.Sprintf("query($login: String!, $cursor: String) { user(login: $login) { %s { %s } } }", connCall, repositoryFields)

	var repos []history.RepositoryObservation
	var cursor string

	for {
		vars := map[string]string{"login": login}
		if cursor != "" {
			vars["cursor"] = cursor
		}
		data, err := client.RunQuery(query, vars)
		if err != nil {
			return nil, err
		}

		userMap, ok := data["user"].(map[string]any)
		if !ok {
			return nil, errors.New("missing user in response")
		}
		connMap, ok := userMap[connName].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("missing connection %s in response", connName)
		}

		nodes, ok := connMap["nodes"].([]any)
		if !ok {
			return nil, errors.New("nodes is not an array")
		}

		for _, node := range nodes {
			nodeMap, ok := node.(map[string]any)
			if !ok {
				continue
			}
			repo, err := parseRepository(nodeMap, languageColors)
			if err != nil {
				return nil, err
			}
			repos = append(repos, repo)
		}

		pageInfo, ok := connMap["pageInfo"].(map[string]any)
		if !ok {
			return nil, errors.New("missing pageInfo")
		}
		hasNext, _ := pageInfo["hasNextPage"].(bool)
		if !hasNext {
			break
		}
		nextCursor, _ := pageInfo["endCursor"].(string)
		if nextCursor == cursor {
			return nil, fmt.Errorf("%s repeated page cursor %s", connName, nextCursor)
		}
		cursor = nextCursor
	}

	return repos, nil
}

func parseRepository(node map[string]any, languageColors map[string]*history.LanguageColor) (history.RepositoryObservation, error) {
	idFloat, _ := node["databaseId"].(float64)
	visStr, _ := node["visibility"].(string)
	isFork, _ := node["isFork"].(bool)

	var vis history.RepositoryVisibility
	switch visStr {
	case "PUBLIC":
		vis = history.VisibilityPublic
	case "PRIVATE":
		vis = history.VisibilityPrivate
	case "INTERNAL":
		vis = history.VisibilityInternal
	default:
		return history.RepositoryObservation{}, fmt.Errorf("unknown repository visibility %s", visStr)
	}

	var primLang *string
	if langObj, ok := node["primaryLanguage"].(map[string]any); ok && langObj != nil {
		langName, _ := langObj["name"].(string)
		if langName != "" {
			primLang = &langName
			var langColor *history.LanguageColor
			if colorStr, ok := langObj["color"].(string); ok && colorStr != "" {
				parsedColor, err := history.ParseLanguageColor(colorStr)
				if err != nil {
					return history.RepositoryObservation{}, err
				}
				langColor = &parsedColor
			}
			if existingColor, exists := languageColors[langName]; exists {
				if (existingColor == nil) != (langColor == nil) || (existingColor != nil && *existingColor != *langColor) {
					return history.RepositoryObservation{}, fmt.Errorf("conflicting language colors for %s", langName)
				}
			} else {
				languageColors[langName] = langColor
			}
		}
	}

	return history.RepositoryObservation{
		RepositoryID:    uint64(idFloat),
		Visibility:      vis,
		IsFork:          isFork,
		PrimaryLanguage: primLang,
	}, nil
}

func collectContributionHistory(client GraphqlClient, login string, createdOn, today history.Date) (*history.ContributionHistory, error) {
	lastCompleteDay := today.AddDays(-1)
	if createdOn.After(lastCompleteDay.Time) {
		return &history.ContributionHistory{
			ObservedOn: lastCompleteDay,
			Days:       nil,
		}, nil
	}

	dailyCounts := make(map[string]uint32)
	windowStart := createdOn

	for !windowStart.After(lastCompleteDay.Time) {
		// Add 12 months, then subtract 1 day
		windowEnd := windowStart.Time.AddDate(0, 12, -1)
		windowEndDate := history.Date{Time: windowEnd}
		if windowEndDate.After(lastCompleteDay.Time) {
			windowEndDate = lastCompleteDay
		}

		windowCounts, err := collectCalendarWindow(client, login, windowStart, windowEndDate)
		if err != nil {
			return nil, err
		}
		for d, count := range windowCounts {
			dailyCounts[d] = count
		}

		windowStart = windowEndDate.AddDays(1)
	}

	var dates []string
	for d := range dailyCounts {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	if len(dates) == 0 || dates[0] != createdOn.String() {
		var firstReported *history.Date
		if len(dates) > 0 {
			d, _ := history.ParseDate(dates[0])
			firstReported = &d
		}
		return nil, fmt.Errorf("incomplete calendar: account created on %s, but first reported day is %v", createdOn, firstReported)
	}

	lastDate, _ := history.ParseDate(dates[len(dates)-1])
	days := make([]history.ContributionDay, len(dates))
	for i, dStr := range dates {
		d, _ := history.ParseDate(dStr)
		days[i] = history.ContributionDay{
			Date:              d,
			ContributionCount: dailyCounts[dStr],
		}
	}

	return &history.ContributionHistory{
		ObservedOn: lastDate,
		Days:       days,
	}, nil
}

func collectCalendarWindow(client GraphqlClient, login string, windowStart, windowEnd history.Date) (map[string]uint32, error) {
	windowCounts := make(map[string]uint32)
	reqEnd := windowEnd

	for {
		reqCounts, earliest, err := requestCalendarDays(client, login, windowStart, reqEnd)
		if err != nil {
			return nil, err
		}
		for d, count := range reqCounts {
			windowCounts[d] = count
		}

		if earliest == nil {
			return windowCounts, nil
		}
		if earliest.Equal(windowStart.Time) {
			return windowCounts, nil
		}
		reqEnd = earliest.AddDays(-1)
	}
}

func requestCalendarDays(client GraphqlClient, login string, start, end history.Date) (map[string]uint32, *history.Date, error) {
	from := fmt.Sprintf("%sT00:00:00Z", start.String())
	to := fmt.Sprintf("%sT23:59:59Z", end.String())

	data, err := client.RunQuery(calendarQuery, map[string]string{
		"login": login,
		"from":  from,
		"to":    to,
	})
	if err != nil {
		return nil, nil, err
	}

	userMap, ok := data["user"].(map[string]any)
	if !ok {
		return nil, nil, errors.New("missing user in calendar query")
	}
	colMap, ok := userMap["contributionsCollection"].(map[string]any)
	if !ok {
		return nil, nil, errors.New("missing contributionsCollection")
	}
	calMap, ok := colMap["contributionCalendar"].(map[string]any)
	if !ok {
		return nil, nil, errors.New("missing contributionCalendar")
	}
	weeks, ok := calMap["weeks"].([]any)
	if !ok {
		return nil, nil, errors.New("weeks is not an array")
	}

	dailyCounts := make(map[string]uint32)
	var earliest *history.Date

	for _, w := range weeks {
		wMap, ok := w.(map[string]any)
		if !ok {
			continue
		}
		days, ok := wMap["contributionDays"].([]any)
		if !ok {
			continue
		}
		for _, day := range days {
			dMap, ok := day.(map[string]any)
			if !ok {
				continue
			}
			dateStr, _ := dMap["date"].(string)
			countFloat, _ := dMap["contributionCount"].(float64)

			d, err := history.ParseDate(dateStr)
			if err != nil {
				return nil, nil, err
			}
			if d.Before(start.Time) || d.After(end.Time) {
				continue
			}

			count := uint32(countFloat)
			if prev, exists := dailyCounts[dateStr]; exists && prev != count {
				return nil, nil, fmt.Errorf("conflicting contribution count for %s", dateStr)
			}
			dailyCounts[dateStr] = count

			if earliest == nil || d.Before(earliest.Time) {
				cp := d
				earliest = &cp
			}
		}
	}

	return dailyCounts, earliest, nil
}
