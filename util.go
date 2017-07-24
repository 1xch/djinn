package djinn

import "fmt"

type settings struct {
	cached bool
}

func defaultSettings() *settings {
	return &settings{
		cached: true,
	}
}

type FuncSet struct {
	f map[string]interface{}
}

func NewFuncSet() *FuncSet {
	return &FuncSet{make(map[string]interface{})}
}

func (f *FuncSet) AddFuncs(fns map[string]interface{}) {
	for k, fn := range fns {
		f.f[k] = fn
	}
}

func (f *FuncSet) GetFuncs() map[string]interface{} {
	return f.f
}

type xrror struct {
	err  string
	vals []interface{}
}

func (x *xrror) Error() string {
	return fmt.Sprintf(x.err, x.vals...)
}

func (x *xrror) Out(vals ...interface{}) *xrror {
	x.vals = vals
	return x
}

func Drror(err string) *xrror {
	return &xrror{err: err}
}

var (
	NilTemplateError   = Drror("nil template named %s").Out
	NoTemplateError    = Drror("no template named %s").Out
	EmptyTemplateError = Drror("empty template named %s").Out
	ConfigurationError = Drror("Could not configure: %s").Out
)
