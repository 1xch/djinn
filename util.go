<<<<<<< HEAD
package jingo
=======
package sweetpl
>>>>>>> ef5537f7c9ae3bc701b1b186db75f22b8f8d4a62

import (
	"fmt"
)

type TemplateError struct {
	Format     string
	Parameters []interface{}
}

func (e *TemplateError) Error() string {
	return fmt.Sprintf(e.Format, e.Parameters...)
}

func Errf(format string, parameters ...interface{}) error {
	return &TemplateError{
		Format:     format,
		Parameters: parameters,
	}
}
