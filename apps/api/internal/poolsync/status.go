package poolsync

import "github.com/FacileStudio/Opus/apps/api/schemas"

const (
	statusTodo       = "to-do"
	statusInProgress = "in-progress"
	statusInReview   = "in-review"
	statusDone       = "done"
	statusArchived   = "archived"
	statusPlanned    = "planned"
)

var sablierRank = map[string]int{statusTodo: 0, statusInProgress: 1, statusInReview: 2}

type target struct {
	slug     string
	columnID *int64
}

func splitColumns(columns []schemas.Column) (open []schemas.Column, final []schemas.Column) {
	for _, column := range columns {
		if column.IsFinal {
			final = append(final, column)
		} else {
			open = append(open, column)
		}
	}
	return open, final
}

func targetOf(column schemas.Column) target {
	id := column.ID
	return target{slug: column.Slug, columnID: &id}
}

func exactColumn(columns []schemas.Column, wire string) (target, bool) {
	for _, column := range columns {
		if column.Slug == wire {
			return targetOf(column), true
		}
	}
	return target{}, false
}

// nearestColumn maps one of Sablier's four statuses onto the project workflow.
func nearestColumn(columns []schemas.Column, wire string) (target, bool) {
	open, final := splitColumns(columns)
	if wire == statusDone {
		if len(final) == 0 {
			return target{}, false
		}
		return targetOf(final[0]), true
	}
	rank, known := sablierRank[wire]
	if !known || len(open) == 0 {
		return target{}, false
	}
	return targetOf(open[min(rank, len(open)-1)]), true
}

// classify reports which of Sablier's statuses a local status stands for.
func classify(columns []schemas.Column, current string) string {
	if _, known := sablierRank[current]; known || current == statusDone {
		return current
	}
	switch current {
	case statusArchived:
		return statusDone
	case statusPlanned:
		return statusTodo
	}
	open, _ := splitColumns(columns)
	for index, column := range open {
		if column.Slug == current {
			return []string{statusTodo, statusInProgress, statusInReview}[min(index, 2)]
		}
	}
	if _, found := exactColumn(columns, current); found {
		return statusDone
	}
	return statusTodo
}

// resolveStatus picks the local status for a wire status. It reports false when
// the wire value is empty, unknown, or already stands for the current status.
func resolveStatus(columns []schemas.Column, wire, current string) (target, bool) {
	if wire == "" {
		return target{}, false
	}
	if exact, found := exactColumn(columns, wire); found {
		return exact, true
	}
	if current != "" && classify(columns, current) == wire {
		return target{}, false
	}
	return nearestColumn(columns, wire)
}

func defaultStatus(columns []schemas.Column) target {
	if open, _ := splitColumns(columns); len(open) > 0 {
		return targetOf(open[0])
	}
	return target{slug: statusTodo}
}
