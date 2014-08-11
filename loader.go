package jingo

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type TemplateLoader interface {
	Load(string) (string, error)
	ListTemplates() interface{}
}

type BaseLoader struct {
	e          []error
	Extensions []string
}

func (b *BaseLoader) ListTemplates() interface{} {
	return "not implemented"
}

func (b *BaseLoader) ValidExtension(ext string) bool {
	for _, extension := range b.Extensions {
		if extension == ext {
			return true
		}
	}
	return false
}

type DirLoader struct {
	BaseLoader
	Paths []string
}

func (l *DirLoader) Load(name string) (string, error) {
	for _, p := range l.Paths {
		f := filepath.Join(p, name)
		if l.ValidExtension(filepath.Ext(f)) {
			if _, err := os.Stat(f); err == nil {
				file, err := os.Open(f)
				r, err := ioutil.ReadAll(file)
				return string(r), err
			}
		}
	}
	return "", Errf("Template %s does not exist", name)
}

func (l *DirLoader) ListTemplates() interface{} {
	return nil
}

func NewDirLoader(basepaths ...string) *DirLoader {
	d := &DirLoader{}
	d.Extensions = append(d.Extensions, ".html", ".jingo")
	for _, p := range basepaths {
		p, err := filepath.Abs(path.Clean(p))
		if err != nil {
			d.e = append(d.e, Errf("path returned error", p))
		}
		d.Paths = append(d.Paths, p)
	}
	return d
}

type MapLoader struct {
	BaseLoader
	m *map[string]string
}

func (l *MapLoader) Load(name string) (string, error) {
	if src, ok := (*l.m)[name]; ok {
		return src, nil
	}
	return "", Errf("Template %s does not exist", name)
}
