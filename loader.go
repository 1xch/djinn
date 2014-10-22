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
		e              []error
		FileExtensions []string
	}

	DirLoader struct {
		BaseLoader
		Paths []string
	}

	MapLoader struct {
		BaseLoader
		m map[string]string
	}
)

func (b *BaseLoader) Load(name string) (string, error) {
	return "not implemented", nil
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
			d.e = append(d.e, DjinnError("path: %s returned error", p))
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
	return "", DjinnError("Template %s does not exist", name)
}

func (l *MapLoader) Load(name string) (string, error) {
	if r, ok := l.m[name]; ok {
		return string(r), nil
	}
	return "", DjinnError("Template %s does not exist", name)
}

func (l *MapLoader) ListTemplates() interface{} {
	var listed []string
	for k, _ := range l.m {
		listed = append(listed, k)
	}
	return listed
}
