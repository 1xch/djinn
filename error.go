package djinn

import (
	"fmt"
)

type djinnError struct {
	err  string
	vals []interface{}
}

func (d *djinnError) Error() string {
	return fmt.Sprintf(d.err, d.vals...)
}

func (d *djinnError) Out(vals ...interface{}) *djinnError {
	d.vals = vals
	return d
}

func Drror(err string) *djinnError {
	return &djinnError{err: err}
}

var (
	NilTemplateError   = Drror("nil template named %s").Out
	NoTemplateError    = Drror("no template named %s").Out
	EmptyTemplateError = Drror("empty template named %s").Out
	ConfigurationError = Drror("Could not configure: %s").Out
)
