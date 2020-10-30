package harvest

import (
	"net/http"
)

const (
	userAgent = "go-harvest"
)

type Client struct {
	client *http.Client

	// User agent used when communicating with the Harvest API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Harvest API.
	TimeEntries *TimeEntriesService
}

type service struct {
	client *Client
}

func NewClient() *Client {
	httpClient := &http.Client{}

	c := &Client{client: httpClient, UserAgent: userAgent}
	c.common.client = c
	c.TimeEntries = (*TimeEntriesService)(&c.common)
	return c
}
