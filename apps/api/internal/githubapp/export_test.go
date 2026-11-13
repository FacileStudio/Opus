package githubapp

import "time"

// SetClock replaces the client's clock in tests.
func SetClock(c *Client, now func() time.Time) {
	c.now = now
}
