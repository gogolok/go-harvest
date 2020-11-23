package harvest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/google/go-querystring/query"
)

const (
	defaultBaseURL    = "https://api.harvestapp.com/"
	defaultApiVersion = "v2/"
	userAgent         = "go-harvest"
)

type Client struct {
	client *http.Client // HTTP client used to communicate with the API.

	// Base URL for API requests. Defaults to the public API.
	// BaseURL should always be specified with a trailing slash.
	BaseURL *url.URL

	AccessToken string // https://help.getharvest.com/api-v2/authentication-api/authentication/authentication/
	AccountId   string // https://help.getharvest.com/api-v2/authentication-api/authentication/authentication/

	// User agent used when communicating with the Harvest API.
	UserAgent string // https://help.getharvest.com/api-v2/authentication-api/authentication/authentication/

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Harvest API.
	TimeEntries *TimeEntriesService
}

type service struct {
	client *Client
}

// ListOptions specifies the optional parameters to various List methods that
// support offset pagination.
type ListOptions struct {
	// For paginated result sets, page of results to retrieve.
	Page int `url:"page,omitempty"`

	// For paginated result sets, the number of results to include per page.
	PerPage int `url:"per_page,omitempty"`
}

// addOptions adds the parameters in opts as URL query parameters to s. opts
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opts interface{}) (string, error) {
	v := reflect.ValueOf(opts)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opts)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func NewClient(accessToken string, accountId string) *Client {
	httpClient := &http.Client{}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL, AccessToken: accessToken, AccountId: accountId, UserAgent: userAgent}
	c.common.client = c
	c.TimeEntries = (*TimeEntriesService)(&c.common)
	return c
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}
	u, err := c.BaseURL.Parse(defaultApiVersion + urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Harvest-Account-Id", c.AccountId)

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is canceled or times out,
// ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	if ctx == nil {
		return nil, errors.New("context must be non-nil")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}

	defer resp.Body.Close()

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)
			if decErr == io.EOF {
				decErr = nil // ignore EOF errors caused by empty response body
			}
			if decErr != nil {
				err = decErr
			}
		}
	}

	return response, err
}

func sanitizeURL(uri *url.URL) *url.URL {
	return uri
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range or equal to 202 Accepted.
// API error responses are expected to have response
// body, and a JSON response body that maps to ErrorResponse.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)

	if err == nil && data != nil {
		json.Unmarshal(data, errorResponse)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	switch {
	case r.StatusCode == http.StatusUnauthorized:
		return (*AuthError)(errorResponse)
	default:
		return errorResponse
	}
}

// An ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response         *http.Response // HTTP response that caused this error
	ErrorCode        string         `json:"error"`             // error code
	ErrorDescription string         `json:"error_description"` // error message
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v %v",
		r.Response.Request.Method, sanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode, r.ErrorCode, r.ErrorDescription)
}

type AuthError ErrorResponse

func (a *AuthError) Error() string {
	r := (*ErrorResponse)(a)
	return fmt.Sprintf("%v %v", r.ErrorCode, r.ErrorDescription)
}

type Pagination struct {
	PerPage      int `json:"per_page"`
	TotalPages   int `json:"total_pages"`
	TotalEntries int `json:"total_entries"`
	NextPage     int `json:"next_page"`
	PreviousPage int `json:"previous_page"`
	Page         int `json:"page"`
}

// Response is a API response. This wraps the standard http.Response
// returned from API and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response
	NextPage     int
	PreviousPage int
	FirstPage    int
	LastPage     int
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	return &Response{Response: r}
}

// populatePageValues parses the HTTP Link response and populates the
// various pagination link values in the Response.
func (r *Response) populatePageValues(pagination Pagination) {
	r.FirstPage = 1
	r.LastPage = pagination.TotalPages

	r.NextPage = pagination.Page + 1
	if r.NextPage > r.LastPage {
		r.NextPage = r.LastPage
	}

	r.PreviousPage = pagination.Page - 1
	if r.PreviousPage < r.FirstPage {
		r.PreviousPage = r.FirstPage
	}
}
