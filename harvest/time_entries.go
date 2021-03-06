package harvest

import (
	"context"
)

type TimeEntriesService service

type Project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Task struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type TimeEntry struct {
	Id        int     `json:"id"`
	Hours     float64 `json:"hours"`
	Notes     string  `json:"notes"`
	Project   Project `json:"project"`
	User      User    `json:"user"`
	Task      Task    `json:"task"`
	SpentDate string  `json:"spent_date"`
}

// TimeEntriesListOptions specifies the optional parameters to the
// TimeEntriesService.List method.
type TimeEntriesListOptions struct {
	From string `url:"from,omitempty"`
	To   string `url:"to,omitempty"`

	ListOptions
}

// List lists time entries.
// https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/#list-all-time-entries
func (t *TimeEntriesService) List(ctx context.Context, opts *TimeEntriesListOptions) ([]*TimeEntry, *Response, error) {
	u := "time_entries"
	u, err := addOptions(u, opts)
	if err != nil {
		return nil, nil, err
	}

	req, err := t.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	type Page struct {
		Pagination
		TimeEntries []*TimeEntry `json:"time_entries"`
	}
	var page Page

	resp, err := t.client.Do(ctx, req, &page)
	if err != nil {
		return nil, resp, err
	}

	resp.populatePageValues(page.Pagination)

	return page.TimeEntries, resp, nil
}
