package harvest

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestTimeEntriesService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/v2/time_entries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `{"time_entries": [{"id": 636709344}]}`)
	})

	pagination_opts := ListOptions{Page: 2}
	opt := &TimeEntriesListOptions{
		ListOptions: pagination_opts,
	}
	timeEntries, _, err := client.TimeEntries.List(context.Background(), opt)
	if err != nil {
		t.Errorf("TimeEntriesService.List returned error: %v", err)
	}

	want := 1
	if len(timeEntries) != want {
		t.Errorf("TimeEntriesService.List returned %+v entries, want %+v", len(timeEntries), want)
	}
}
