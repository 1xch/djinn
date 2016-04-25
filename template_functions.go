package djinn

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
