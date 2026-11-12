package collect_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/reginaldalfret/git-koi-aquarium/pkg/collect"
	"github.com/reginaldalfret/git-koi-aquarium/pkg/history"
)

type mockGraphqlClient struct {
	handlers map[string]func(vars map[string]string) (map[string]any, error)
}

func (m *mockGraphqlClient) RunQuery(query string, variables map[string]string) (map[string]any, error) {
	for prefix, handler := range m.handlers {
		if strings.Contains(query, prefix) {
			return handler(variables)
		}
	}
	return nil, errors.New("unhandled mock query: " + query)
}

func TestCollectAccountSnapshotSuccess(t *testing.T) {
	today := history.NewDate(2026, 11, 10)
	createdOn := history.NewDate(2026, 11, 8)

	mock := &mockGraphqlClient{
		handlers: map[string]func(vars map[string]string) (map[string]any, error){
			"databaseId createdAt": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"databaseId": float64(12345),
						"createdAt":  "2026-11-08T12:00:00Z",
					},
				}, nil
			},
			"repositories(": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"repositories": map[string]any{
							"nodes": []any{
								map[string]any{
									"databaseId": float64(101),
									"visibility": "PUBLIC",
									"isFork":     false,
									"primaryLanguage": map[string]any{
										"name":  "Go",
										"color": "#00add8",
									},
								},
							},
							"pageInfo": map[string]any{
								"hasNextPage": false,
							},
						},
					},
				}, nil
			},
			"repositoriesContributedTo(": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"repositoriesContributedTo": map[string]any{
							"nodes": []any{},
							"pageInfo": map[string]any{
								"hasNextPage": false,
							},
						},
					},
				}, nil
			},
			"contributionsCollection(": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"contributionsCollection": map[string]any{
							"contributionCalendar": map[string]any{
								"weeks": []any{
									map[string]any{
										"contributionDays": []any{
											map[string]any{"date": "2026-11-08", "contributionCount": float64(3)},
											map[string]any{"date": "2026-11-09", "contributionCount": float64(5)},
										},
									},
								},
							},
						},
					},
				}, nil
			},
		},
	}

	snapshot, err := collect.CollectAccountSnapshot(mock, "octocat", today)
	if err != nil {
		t.Fatalf("CollectAccountSnapshot failed: %v", err)
	}

	if snapshot.AccountID != 12345 {
		t.Errorf("expected account ID 12345, got %d", snapshot.AccountID)
	}
	if len(snapshot.Repositories) != 1 {
		t.Errorf("expected 1 repository, got %d", len(snapshot.Repositories))
	}
	if snapshot.ContributionHistory == nil || len(snapshot.ContributionHistory.Days) != 2 {
		t.Errorf("expected 2 contribution days, got %v", snapshot.ContributionHistory)
	}
	if snapshot.LanguageColors["Go"] == nil || snapshot.LanguageColors["Go"].Hex() != "#00add8" {
		t.Errorf("expected #00add8 for Go, got %v", snapshot.LanguageColors["Go"])
	}
	_ = createdOn
}

func TestCollectAccountNotFound(t *testing.T) {
	mock := &mockGraphqlClient{
		handlers: map[string]func(vars map[string]string) (map[string]any, error){
			"databaseId createdAt": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": nil,
				}, nil
			},
		},
	}
	today := history.NewDate(2026, 11, 10)
	_, err := collect.CollectAccountSnapshot(mock, "nonexistent-user", today)
	if err == nil {
		t.Errorf("expected error for non-existent user, got nil")
	}
}

func TestCollectRepeatedCursorPagination(t *testing.T) {
	mock := &mockGraphqlClient{
		handlers: map[string]func(vars map[string]string) (map[string]any, error){
			"databaseId createdAt": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"databaseId": float64(123),
						"createdAt":  "2026-11-08T00:00:00Z",
					},
				}, nil
			},
			"repositories(": func(vars map[string]string) (map[string]any, error) {
				return map[string]any{
					"user": map[string]any{
						"repositories": map[string]any{
							"nodes": []any{},
							"pageInfo": map[string]any{
								"hasNextPage": true,
								"endCursor":   "same_cursor",
							},
						},
					},
				}, nil
			},
		},
	}
	today := history.NewDate(2026, 11, 10)
	_, err := collect.CollectAccountSnapshot(mock, "looping-user", today)
	if err == nil || !strings.Contains(err.Error(), "repeated page cursor") {
		t.Errorf("expected repeated page cursor error, got %v", err)
	}
}

func TestCollectGraphQLError(t *testing.T) {
	mock := &mockGraphqlClient{
		handlers: map[string]func(vars map[string]string) (map[string]any, error){
			"databaseId createdAt": func(vars map[string]string) (map[string]any, error) {
				return nil, errors.New("rate limit exceeded")
			},
		},
	}
	today := history.NewDate(2026, 11, 10)
	_, err := collect.CollectAccountSnapshot(mock, "ratelimited-user", today)
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected rate limit error, got %v", err)
	}
}
