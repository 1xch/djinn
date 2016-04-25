package djinn

import (
	"sort"
)

type ConfigFn func(*Djinn) error

type Config interface {
	Order() int
	Configure(*Djinn) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(d *Djinn) error {
	return c.fn(d)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

type Configuration interface {
	AddConfig(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	d          *Djinn
	configured bool
	list       configList
}

func newConfiguration(d *Djinn, conf ...Config) *configuration {
	c := &configuration{
		d:    d,
		list: builtIns,
	}
	c.AddConfig(conf...)
	return c
}

func (c *configuration) AddConfig(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(d *Djinn, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(d)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.d, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1000, setCache},
}

func setCache(d *Djinn) error {
	if d.cached {
		if d.Cache == nil {
			d.Cache = TLRUCache(100)
		}
	}
	return nil
}

func CacheOn(c Cache) Config {
	return DefaultConfig(
		func(d *Djinn) error {
			d.Cache = c
			d.cached = true
			return nil
		})
}

func Loaders(l ...Loader) Config {
	return DefaultConfig(func(d *Djinn) error {
		d.AddLoaders(l...)
		return nil
	})
}

func TemplateFunctions(f ...map[string]interface{}) Config {
	return DefaultConfig(func(d *Djinn) error {
		for _, ff := range f {
			d.AddFuncs(ff)
		}
		return nil
	})
}
