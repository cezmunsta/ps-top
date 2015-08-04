// Package file_io_latency holds the routines which manage the file_summary_by_instance table.
package file_io_latency

import (
	"database/sql"
	"fmt"

	"github.com/sjmudd/ps-top/baseobject"
)

// Object represents the contents of the data collected from file_summary_by_instance
type Object struct {
	baseobject.BaseObject // embedded
	initial               Rows
	current               Rows
	results               Rows
	totals                Row
	globalVariables       map[string]string
}

// SetInitialFromCurrent resets the statistics to current values
func (t *Object) SetInitialFromCurrent() {
	t.initial = make(Rows, len(t.current))
	copy(t.initial, t.current)

	t.results = make(Rows, len(t.current))
	copy(t.results, t.current)

	if t.WantRelativeStats() {
		t.results.subtract(t.initial) // should be 0 if relative
	}

	t.results.sort()
	t.totals = t.results.totals()
}

// Collect data from the db, then merge it in.
func (t *Object) Collect(dbh *sql.DB) {
	// UPDATE current from db handle
	t.current = mergeByTableName(selectRows(dbh), t.globalVariables)
	t.SetNow()

	// copy in initial data if it was not there
	if len(t.initial) == 0 && len(t.current) > 0 {
		t.initial = make(Rows, len(t.current))
		copy(t.initial, t.current)
	}

	// check for reload initial characteristics
	if t.initial.needsRefresh(t.current) {
		t.initial = make(Rows, len(t.current))
		copy(t.initial, t.current)
	}

	t.makeResults()
}

func (t *Object) makeResults() {
	t.results = make(Rows, len(t.current))
	copy(t.results, t.current)
	if t.WantRelativeStats() {
                t.results.subtract(t.initial)
	}

	t.results.sort()
	t.totals = t.results.totals()
}

// Headings returns the headings for a table
func (t Object) Headings() string {
	var r Row

	return r.headings()
}

// RowContent returns the rows we need for displaying
func (t Object) RowContent(maxRows int) []string {
	rows := make([]string, 0, maxRows)

	for i := range t.results {
		if i < maxRows {
			rows = append(rows, t.results[i].rowContent(t.totals))
		}
	}

	return rows
}

// Len return the length of the result set
func (t Object) Len() int {
	return len(t.results)
}

// TotalRowContent returns all the totals
func (t Object) TotalRowContent() string {
	return t.totals.rowContent(t.totals)
}

// EmptyRowContent returns an empty string of data (for filling in)
func (t Object) EmptyRowContent() string {
	var empty Row
	return empty.rowContent(empty)
}

// Description returns a description of the table
func (t Object) Description() string {
	var count int
	for row := range t.results {
		if t.results[row].sumTimerWait > 0 {
			count++
		}
	}

	return fmt.Sprintf("File I/O Latency (file_summary_by_instance) %4d row(s)    ", count)
}

// NewFileSummaryByInstance creates a new structure and include various variable values:
// - datadir, relay_log
// There's no checking that these are actually provided!
func NewFileSummaryByInstance(globalVariables map[string]string) *Object {
	n := new(Object)

	n.globalVariables = globalVariables

	return n
}
