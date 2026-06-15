package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/SuprPhatAnon/phatodo/internal/domain"
)

func toonIndent(level int) string {
	return strings.Repeat("  ", level)
}

func toonScalar(value string) string {
	if value == "" {
		return `""`
	}
	if needsTOONQuotes(value) {
		return strconv.Quote(value)
	}
	return value
}

func needsTOONQuotes(value string) bool {
	for _, r := range value {
		switch r {
		case ' ', '\t', '\n', '\r', ':', ',', '{', '}', '[', ']', '"', '#':
			return true
		}
	}
	return strings.HasPrefix(value, "-") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ")
}

func writeTOONField(w io.Writer, indent int, key string, value string) {
	fmt.Fprintf(w, "%s%s: %s\n", toonIndent(indent), key, toonScalar(value))
}

func writeTOONQuotedField(w io.Writer, indent int, key string, value string) {
	fmt.Fprintf(w, "%s%s: %s\n", toonIndent(indent), key, strconv.Quote(value))
}

func writeTOONIntField(w io.Writer, indent int, key string, value int) {
	fmt.Fprintf(w, "%s%s: %d\n", toonIndent(indent), key, value)
}

func writeTOONArrayHeader(w io.Writer, indent int, key string, count int) {
	fmt.Fprintf(w, "%s%s[%d]:\n", toonIndent(indent), key, count)
}

func writeTOONArrayHeaderWithFields(w io.Writer, indent int, key string, count int, fields []string) {
	fmt.Fprintf(w, "%s%s[%d]{%s}:\n", toonIndent(indent), key, count, strings.Join(fields, ","))
}

func writeTOONListItemStart(w io.Writer, indent int, key string, value string) {
	fmt.Fprintf(w, "%s- %s: %s\n", toonIndent(indent), key, toonScalar(value))
}

func writeTOONRowValue(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, ",:\n\r\t\"") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		return strconv.Quote(value)
	}
	return value
}

func writeTOONRow(w io.Writer, indent int, values ...string) {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = writeTOONRowValue(value)
	}
	fmt.Fprintf(w, "%s- %s\n", toonIndent(indent), strings.Join(parts, ","))
}

func writeTOONStringArray(w io.Writer, indent int, key string, values []string) {
	writeTOONArrayHeader(w, indent, key, len(values))
	for _, value := range values {
		fmt.Fprintf(w, "%s- %s\n", toonIndent(indent+1), toonScalar(value))
	}
}

func writeProjectConfigItem(w io.Writer, item ProjectConfigItem) {
	writeTOONListItemStart(w, 0, item.Key, item.Value)
}

func writeTaskCreateResponse(w io.Writer, resp domain.TaskCreateResponse) {
	writeTOONListItemStart(w, 0, "id", resp.ID)
	writeTOONField(w, 1, "issue_prefix", resp.IssuePrefix)
	writeTOONField(w, 1, "title", resp.Title)
	writeTOONField(w, 1, "status", string(resp.Status))
	writeTOONIntField(w, 1, "priority", int(resp.Priority))
	writeTOONField(w, 1, "project_id", resp.ProjectID)
	writeTOONField(w, 1, "workspace_id", resp.WorkspaceID)
}

func writeTOONTimeField(w io.Writer, indent int, key string, value time.Time) {
	if value.IsZero() {
		return
	}
	writeTOONQuotedField(w, indent, key, value.UTC().Format(time.RFC3339Nano))
}

func writeTaskListItem(w io.Writer, indent int, item domain.TaskListItem) {
	writeTOONListItemStart(w, indent, "id", item.ID)
	writeTOONField(w, indent+1, "title", item.Title)
	writeTOONField(w, indent+1, "description", item.Description)
	writeTOONIntField(w, indent+1, "priority", int(item.Priority))
	writeTOONField(w, indent+1, "status", string(item.Status))
	if item.EpicID != "" {
		writeTOONField(w, indent+1, "epicId", item.EpicID)
	}
	if item.ParentTaskID != "" {
		writeTOONField(w, indent+1, "parentTaskId", item.ParentTaskID)
	}
	if len(item.Tags) > 0 {
		writeTOONQuotedField(w, indent+1, "tags", strings.Join(item.Tags, ","))
	}
	writeTOONTimeField(w, indent+1, "createdAt", item.CreatedAt)
	writeTOONTimeField(w, indent+1, "updatedAt", item.UpdatedAt)
}

func writeReadyListItem(w io.Writer, indent int, item domain.ReadyListItem) {
	writeTOONListItemStart(w, indent, "id", item.ID)
	writeTOONField(w, indent+1, "title", item.Title)
	writeTOONField(w, indent+1, "description", item.Description)
	writeTOONIntField(w, indent+1, "priority", int(item.Priority))
	writeTOONField(w, indent+1, "status", string(item.Status))
	if item.EpicID != "" {
		writeTOONField(w, indent+1, "epicId", item.EpicID)
	}
	if item.ParentTaskID != "" {
		writeTOONField(w, indent+1, "parentTaskId", item.ParentTaskID)
	}
	if len(item.Tags) > 0 {
		writeTOONQuotedField(w, indent+1, "tags", strings.Join(item.Tags, ","))
	}
	writeTOONTimeField(w, indent+1, "createdAt", item.CreatedAt)
	writeTOONTimeField(w, indent+1, "updatedAt", item.UpdatedAt)
	if len(item.Unblocks) > 0 {
		writeTOONArrayHeaderWithFields(w, indent+1, "dependents", len(item.Unblocks), []string{"id", "title", "status", "priority"})
		for _, blocked := range item.Unblocks {
			writeTOONRow(w, indent+2, blocked.ID, blocked.Title, string(blocked.Status), strconv.Itoa(int(blocked.Priority)))
		}
	}
}

func writeTaskDetail(w io.Writer, item domain.TaskDetail) {
	writeTOONListItemStart(w, 0, "id", item.ID)
	writeTOONField(w, 1, "title", item.Title)
	writeTOONField(w, 1, "description", item.Description)
	writeTOONIntField(w, 1, "priority", int(item.Priority))
	writeTOONField(w, 1, "status", string(item.Status))
	if item.EpicID != "" {
		writeTOONField(w, 1, "epicId", item.EpicID)
	}
	if item.ParentTaskID != "" {
		writeTOONField(w, 1, "parentTaskId", item.ParentTaskID)
	}
	if item.AssignedTo != "" {
		writeTOONField(w, 1, "assignedTo", item.AssignedTo)
	}
	if item.CreatedBy != "" {
		writeTOONField(w, 1, "createdBy", item.CreatedBy)
	}
	if item.UpdatedBy != "" {
		writeTOONField(w, 1, "updatedBy", item.UpdatedBy)
	}
	if item.CompletedBy != "" {
		writeTOONField(w, 1, "completedBy", item.CompletedBy)
	}
	if len(item.Tags) > 0 {
		writeTOONQuotedField(w, 1, "tags", strings.Join(item.Tags, ","))
	}
	if len(item.AcceptanceCriteria) > 0 {
		writeTOONStringArray(w, 1, "acceptanceCriteria", item.AcceptanceCriteria)
	}
	if len(item.CompletionEvidence) > 0 {
		writeTOONStringArray(w, 1, "completionEvidence", item.CompletionEvidence)
	}
	if item.CompletionSummary != "" {
		writeTOONField(w, 1, "completionSummary", item.CompletionSummary)
	}
	writeTOONTimeField(w, 1, "completedAt", item.CompletedAt)
	writeTOONTimeField(w, 1, "createdAt", item.CreatedAt)
	writeTOONTimeField(w, 1, "updatedAt", item.UpdatedAt)
}

func writeComment(w io.Writer, indent int, item domain.Comment) {
	writeTOONListItemStart(w, indent, "id", item.ID)
	writeTOONField(w, indent+1, "author", item.Author)
	writeTOONField(w, indent+1, "kind", item.Kind)
	writeTOONField(w, indent+1, "content", item.Content)
	writeTOONTimeField(w, indent+1, "createdAt", item.CreatedAt)
	writeTOONTimeField(w, indent+1, "updatedAt", item.UpdatedAt)
}
