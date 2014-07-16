<<<<<<< HEAD
package jingo
=======
package sweetpl
>>>>>>> ef5537f7c9ae3bc701b1b186db75f22b8f8d4a62

import (
	"io/ioutil"
	"os"
)

type TemplateLoader interface {
	LoadTemplate(string) (string, error)
}

type DirLoader struct {
	BasePath string
}

// DirLoader.LoadTemplate gets BaseAddress + name. No safety checking yet.
func (l *DirLoader) LoadTemplate(name string) (string, error) {
	file, err := os.Open(l.BasePath + "/" + name)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(file)
	return string(b), err
}

type MapLoader map[string]string

// MapLoader.LoadTemplate gets name from a map.
func (l *MapLoader) LoadTemplate(name string) (string, error) {
	src, ok := (*l)[name]
	if !ok {
		return "", Errf("Could not find template " + name)
	}
	return src, nil
<<<<<<< HEAD
=======
	//buff := bytes.NewBufferString(src)
	//return buff, nil
>>>>>>> ef5537f7c9ae3bc701b1b186db75f22b8f8d4a62
}
