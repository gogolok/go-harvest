package harvest

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient()

	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}
}
