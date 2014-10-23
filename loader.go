package djinn

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type (
	TemplateLoader interface {
		Load(string) (string, error)
		ListTemplates() interface{}
	}

	BaseLoader struct {
		Errors         []error
		FileExtensions []string
	}

	DirLoader struct {
		BaseLoader
		Paths []string
	}

	MapLoader struct {
		BaseLoader
		TemplateMap map[string]string
	}
)

func (b *BaseLoader) Load(name string) (string, error) {
	return "", DjinnError("load method not implemented")
}

func (b *BaseLoader) ListTemplates() interface{} {
	return "not implemented"
}

func (b *BaseLoader) ValidExtension(ext string) bool {
	for _, extension := range b.FileExtensions {
		if extension == ext {
			return true
		}
	}
	return false
}

func NewDirLoader(basepaths ...string) *DirLoader {
	d := &DirLoader{}
	d.FileExtensions = append(d.FileExtensions, ".html", ".dji")
	for _, p := range basepaths {
		p, err := filepath.Abs(filepath.Clean(p))
		if err != nil {
			d.Errors = append(d.Errors, DjinnError("path: %s returned error", p))
		}
		d.Paths = append(d.Paths, p)
	}
	return d
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
	return "", DjinnError("template %s does not exist", name)
}

func (l *DirLoader) ListTemplates() interface{} {
	var listing []string
	for _, p := range l.Paths {
		filepath.Walk(p, func(path string, _ os.FileInfo, _ error) (err error) {
			tem := filepath.Base(path)
			if l.ValidExtension(filepath.Ext(tem)) {
				listing = append(listing, tem)
			}
			return err
		})
	}
	return listing
}

func NewMapLoader(tm ...map[string]string) *MapLoader {
	m := &MapLoader{TemplateMap: make(map[string]string)}
	for _, t := range tm {
		for k, v := range t {
			m.TemplateMap[k] = v
		}
	}
	return m
}

func (l *MapLoader) Load(name string) (string, error) {
	if r, ok := l.TemplateMap[name]; ok {
		return string(r), nil
	}
	return "", DjinnError("template %s does not exist", name)
}

func (l *MapLoader) ListTemplates() interface{} {
	var listing []string
	for k, _ := range l.TemplateMap {
		listing = append(listing, k)
	}
	return listing
}
