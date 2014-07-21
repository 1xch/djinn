package jingo

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type TemplateLoader interface {
	LoadTemplate(string) (string, error)
	ListTemplates() interface{}
}

type BaseLoader struct {
	e error
}

func (b *BaseLoader) ListTemplates() interface{} {
	return "not implemented"
}

type DirLoader struct {
	BaseLoader
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

func (l *DirLoader) ListTemplates() interface{} {
	return nil
}

func NewDirLoader(basepath string) *DirLoader {
	d := &DirLoader{}
	b, err := filepath.Abs(basepath)
	if err != nil {
		d.e = Errf("basepath returned error", basepath)
	}
	d.BasePath = b
	return d
}

type MapLoader struct {
	BaseLoader
	m *map[string]string
}

// MapLoader.LoadTemplate gets name from a map.
func (l *MapLoader) LoadTemplate(name string) (string, error) {
	src, ok := (*l.m)[name]
	if !ok {
		return "", Errf("Could not find template " + name)
	}
	return src, nil
}
