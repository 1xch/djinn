package djinn

import "reflect"

type (
	// A configuration function that takes an Djinn pointer, configures the
	// *Djinn within the function, and returning an error.
	Conf func(*Djinn) error

	conf struct {
		CacheOn bool
	}
)

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

func (j *Djinn) elem() reflect.Value {
	v := reflect.ValueOf(j)
	return v.Elem()
}

func (j *Djinn) getfield(fieldname string) reflect.Value {
	return j.elem().FieldByName(fieldname)
}

func (j *Djinn) SetConfBool(fieldname string, as bool) error {
	f := j.getfield(fieldname)
	if f.CanSet() {
		f.SetBool(as)
		return nil
	}
	return DjinnError("could not set field %s as %t", fieldname, as)
}
