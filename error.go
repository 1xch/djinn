package djinn

import (
	"fmt"
)

type djinnerror struct {
	Format     string
	Parameters []interface{}
}

func (e *djinnerror) Error() string {
	return fmt.Sprintf(e.Format, e.Parameters...)
}

func DjinnError(format string, parameters ...interface{}) error {
	return &djinnerror{
		Format:     format,
		Parameters: parameters,
	}
}
