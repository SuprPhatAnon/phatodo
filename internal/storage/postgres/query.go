package postgres

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
	db "github.com/SuprPhatAnon/phatodo/internal/storage/postgres/sqlc"
)

type sortTerm struct {
	field string
	desc  bool
}

func (s *Store) Search(ctx context.Context, projectID string, query string, entityType string, status string, limit int) (domain.SearchResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.SearchResponse{}, err
	}

	types := splitCSV(entityType)
	if len(types) == 0 {
		types = []string{"epic", "task", "subtask", "comment"}
	}
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}

	items := make([]domain.SearchItem, 0)
	if typeSet["epic"] {
		rows, err := s.q.SearchEpics(ctx, db.SearchEpicsParams{ProjectID: projectID, Status: status, Query: query})
		if err != nil {
			return domain.SearchResponse{}, fmt.Errorf("query epic search: %w", err)
		}
		for _, row := range rows {
			items = append(items, searchItemFromEpicSQLC(row))
		}
	}

	if typeSet["task"] || typeSet["subtask"] {
		rows, err := s.q.SearchTasks(ctx, db.SearchTasksParams{ProjectID: projectID, Status: status, Query: query})
		if err != nil {
			return domain.SearchResponse{}, fmt.Errorf("query task search: %w", err)
		}
		for _, row := range rows {
			item := searchItemFromTaskSQLC(row)
			if item.EntityType == "subtask" && !typeSet["subtask"] {
				continue
			}
			if item.EntityType == "task" && !typeSet["task"] {
				continue
			}
			items = append(items, item)
		}
	}

	if typeSet["comment"] {
		rows, err := s.q.SearchComments(ctx, db.SearchCommentsParams{ProjectID: projectID, Query: query})
		if err != nil {
			return domain.SearchResponse{}, fmt.Errorf("query comment search: %w", err)
		}
		for _, row := range rows {
			items = append(items, searchItemFromCommentSQLC(row))
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
			return items[i].CreatedAt.After(items[j].CreatedAt)
		}
		if items[i].EntityType != items[j].EntityType {
			return items[i].EntityType < items[j].EntityType
		}
		return items[i].ID < items[j].ID
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return domain.SearchResponse{
		ProjectID: projectID,
		Query:     query,
		Items:     items,
	}, nil
}

func (s *Store) ListUnified(ctx context.Context, projectID string, entityType string, status string, priority string, sortSpec string, limit int) (domain.ListResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.ListResponse{}, err
	}

	types := splitCSV(entityType)
	if len(types) == 0 {
		types = []string{"epic", "task", "subtask"}
	}
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	prioritySet := parsePrioritySet(priority)

	items := make([]domain.UnifiedListItem, 0)
	if typeSet["epic"] {
		rows, err := s.q.ListEpics(ctx, db.ListEpicsParams{ProjectID: projectID, Status: status})
		if err != nil {
			return domain.ListResponse{}, fmt.Errorf("query epic list: %w", err)
		}
		for _, row := range rows {
			items = append(items, unifiedItemFromEpicSQLC(row))
		}
	}

	if typeSet["task"] || typeSet["subtask"] {
		rows, err := s.q.ListTasksUnified(ctx, db.ListTasksUnifiedParams{ProjectID: projectID, Status: status})
		if err != nil {
			return domain.ListResponse{}, fmt.Errorf("query task list: %w", err)
		}
		for _, row := range rows {
			item := unifiedItemFromTaskSQLC(row)
			if item.EntityType == "subtask" && !typeSet["subtask"] {
				continue
			}
			if item.EntityType == "task" && !typeSet["task"] {
				continue
			}
			if len(prioritySet) > 0 && !prioritySet[int(item.Priority)] {
				continue
			}
			items = append(items, item)
		}
	}

	sortUnifiedItems(items, sortSpec)
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return domain.ListResponse{
		ProjectID: projectID,
		Items:     items,
	}, nil
}

func (s *Store) History(ctx context.Context, projectID string, entityID string, entityType string, action string, since string, limit int) (domain.HistoryResponse, error) {
	if err := s.ensureProjectExists(ctx, projectID); err != nil {
		return domain.HistoryResponse{}, err
	}

	queryLimit := int64(limit)
	if queryLimit <= 0 {
		queryLimit = 2147483647
	}

	var rows []db.Event
	var err error
	if since == "" {
		rows, err = s.q.HistoryEvents(ctx, db.HistoryEventsParams{
			ProjectID:  projectID,
			EntityID:   entityID,
			EntityType: entityType,
			Action:     action,
			QueryLimit: queryLimit,
		})
	} else {
		sinceTime, parseErr := parseFlexibleTime(since)
		if parseErr != nil {
			return domain.HistoryResponse{}, parseErr
		}
		rows, err = s.q.HistoryEventsSince(ctx, db.HistoryEventsSinceParams{
			ProjectID:  projectID,
			EntityID:   entityID,
			EntityType: entityType,
			Action:     action,
			Since:      sinceTime,
			QueryLimit: queryLimit,
		})
	}
	if err != nil {
		return domain.HistoryResponse{}, fmt.Errorf("query history: %w", err)
	}

	items := make([]domain.HistoryEvent, 0)
	for _, row := range rows {
		items = append(items, historyEventFromSQLC(row))
	}

	return domain.HistoryResponse{
		ProjectID: projectID,
		Items:     items,
	}, nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func parsePrioritySet(value string) map[int]bool {
	values := splitCSV(value)
	if len(values) == 0 {
		return nil
	}
	result := make(map[int]bool, len(values))
	for _, v := range values {
		if n, err := strconv.Atoi(v); err == nil {
			result[n] = true
		}
	}
	return result
}

func parseFlexibleTime(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid since value: %s", value)
}

func sortUnifiedItems(items []domain.UnifiedListItem, sortSpec string) {
	terms := parseSortSpec(sortSpec)
	sort.Slice(items, func(i, j int) bool {
		for _, term := range terms {
			switch term.field {
			case "priority":
				if items[i].Priority != items[j].Priority {
					if term.desc {
						return items[i].Priority > items[j].Priority
					}
					return items[i].Priority < items[j].Priority
				}
			case "created":
				if !items[i].CreatedAt.Equal(items[j].CreatedAt) {
					if term.desc {
						return items[i].CreatedAt.After(items[j].CreatedAt)
					}
					return items[i].CreatedAt.Before(items[j].CreatedAt)
				}
			case "updated":
				if !items[i].UpdatedAt.Equal(items[j].UpdatedAt) {
					if term.desc {
						return items[i].UpdatedAt.After(items[j].UpdatedAt)
					}
					return items[i].UpdatedAt.Before(items[j].UpdatedAt)
				}
			case "type":
				if items[i].EntityType != items[j].EntityType {
					if term.desc {
						return items[i].EntityType > items[j].EntityType
					}
					return items[i].EntityType < items[j].EntityType
				}
			case "title":
				if items[i].Title != items[j].Title {
					if term.desc {
						return items[i].Title > items[j].Title
					}
					return items[i].Title < items[j].Title
				}
			case "id":
				if items[i].ID != items[j].ID {
					if term.desc {
						return items[i].ID > items[j].ID
					}
					return items[i].ID < items[j].ID
				}
			}
		}
		return items[i].ID < items[j].ID
	})
}

func parseSortSpec(value string) []sortTerm {
	value = strings.TrimSpace(value)
	if value == "" {
		return []sortTerm{{field: "priority"}, {field: "created"}}
	}
	parts := strings.Split(value, ",")
	terms := make([]sortTerm, 0, len(parts))
	for _, part := range parts {
		segments := strings.SplitN(strings.TrimSpace(part), ":", 2)
		term := sortTerm{field: strings.TrimSpace(segments[0])}
		if len(segments) == 2 && strings.EqualFold(strings.TrimSpace(segments[1]), "desc") {
			term.desc = true
		}
		if term.field != "" {
			terms = append(terms, term)
		}
	}
	if len(terms) == 0 {
		return []sortTerm{{field: "priority"}, {field: "created"}}
	}
	return terms
}
