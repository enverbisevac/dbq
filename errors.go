package dbq

import "fmt"

type NotFoundError struct {
	DataSource string
}

func (e NotFoundError) Error() string {
	if e.DataSource != "" {
		return fmt.Sprintf("no data found in %v", e.DataSource)
	}
	return "no data found"
}
