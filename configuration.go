package djinn

import "reflect"

var ConfigurationError = Drror("Could not configure: %s").Out

// Any function taking a Djinn pointer for configuration.
type Conf func(*Djinn) error

type conf struct {
	CacheOn bool
}

func defaultconf() *conf {
	return &conf{
		CacheOn: false,
	}
}

// SetConf takes any number of Conf functions to configure the Djinn.
func (j *Djinn) SetConf(opts ...Conf) error {
	for _, opt := range opts {
		if err := opt(j); err != nil {
			return err
		}
	}
	return nil
}

// Cache is a Conf function that sets the Djinn cache on, using the provided Cache.
func CacheOn(c Cache) Conf {
	return func(j *Djinn) error {
		j.Cache = c
		return j.SetConfBool("CacheOn", true)
	}
}

func (j *Djinn) addloaders(loaders ...TemplateLoader) {
	for _, l := range loaders {
		j.Loaders = append(j.Loaders, l)
	}
}

// Loaders is a Conf function that adds template loaders to the Djinn.
func Loaders(loaders ...TemplateLoader) Conf {
	return func(j *Djinn) error {
		j.addloaders(loaders...)
		return nil
	}
}

func (j *Djinn) addfunctions(f ...map[string]interface{}) {
	for _, m := range f {
		for k, v := range m {
			j.FuncMap[k] = v
		}
	}
}

// TemplateFunctions is a Conf function adding template functions through any
// number of string-interface maps.
func TemplateFunctions(f ...map[string]interface{}) Conf {
	return func(j *Djinn) error {
		j.addfunctions(f...)
		return nil
	}
}

func (j *Djinn) elem() reflect.Value {
	v := reflect.ValueOf(j)
	return v.Elem()
}

func (j *Djinn) getfield(fieldname string) reflect.Value {
	return j.elem().FieldByName(fieldname)
}

var FieldSetError = Drror("could not set field %s as %t").Out

func (j *Djinn) SetConfBool(fieldname string, as bool) error {
	f := j.getfield(fieldname)
	if f.CanSet() {
		f.SetBool(as)
		return nil
	}
	return FieldSetError(fieldname, as)
}
