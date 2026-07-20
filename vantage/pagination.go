package vantage

import (
	"fmt"
	"net/url"
	"strconv"
)

// pageFromURL extracts the "page" query parameter from a pagination link URL.
func pageFromURL(rawURL string) (int32, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, err
	}
	pageStr := u.Query().Get("page")
	if pageStr == "" {
		return 0, fmt.Errorf("no page parameter in URL")
	}
	n, err := strconv.ParseInt(pageStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("page parameter %q is not an integer: %w", pageStr, err)
	}
	return int32(n), nil
}
