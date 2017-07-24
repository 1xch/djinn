package djinn

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

var m1 map[string]string = map[string]string{
	"vars.html":                      `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
	"Folder/Main-file_name.html.dji": tplMain,
	"sub1.html":                      tplSub1,
	"sub2.html":                      tplSub2,
	"plaintext.html":                 "<Plain>",
}

var m2 map[string]string = map[string]string{
	"varsa.html":                      `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
	"Folder/Main-file_namea.html.dji": tplMain,
	"sub1a.html":                      tplSub1,
	"sub2a.html":                      tplSub2a,
	"plaintext.html":                  "<Plain>",
}

var J1 *Djinn = New(Loaders(MapLoader(m1), MapLoader(m2)))

var J2 *Djinn = New(Loaders(DirLoader("./test/templates"), DirLoader("./test/additional/templates")))

var J3 *Djinn = New(Loaders(MapLoader(m1), DirLoader("./test/additional/templates")))

type TemplateData struct {
	Title string
	Data  map[string]interface{}
}

func MustContain(t *testing.T, str string, check string) {
	index := strings.Index(str, check)
	if index == -1 {
		t.Errorf("Template did not contain %s in the correct order", check)
	}
}

func tplRun(t *testing.T, j *Djinn, name string, data interface{}, check ...string) {
	w := &bytes.Buffer{}

	err := j.Render(w, name, data)

	// fetch
	if _, err = j.Fetch(name); err != nil {
		t.Errorf("Could not fetch %s from %+v", name, j)
	}

	// from cache
	j.Render(w, name, data)

	for _, c := range check {
		MustContain(t, w.String(), c)
	}

}

func TestTemplate(t *testing.T) {
	data := &TemplateData{
		Title: "Hello World",
		Data: map[string]interface{}{
			"Key":   "Value",
			"Slice": []string{"One", "Two", "Three"},
		},
	}

	data1 := &TemplateData{
		Title: "Hello World A",
		Data: map[string]interface{}{
			"Key":   "Value A",
			"Slice": []string{"A", "B", "C"},
		},
	}

	js := []*Djinn{
		J1,
		J2,
		J3,
	}

	for _, cj := range js {
		err := cj.Configure()
		if err != nil {
			t.Errorf("configuration error: %s", err.Error())
		}
	}

	for _, j := range js {
		tplRun(t, j, "vars.html", data, "<title>Hello World</title>", "Key=Value")
		tplRun(t, j, "sub2.html", data, "<MAIN>", "<SUB1>", "<SUB2>", "</SUB2>", "</SUB1>", "</MAIN>")
		tplRun(t, j, "sub1.html", data, "<MAIN>", "<SUB1>", "</SUB1>", "<Plain>", "</MAIN>")
		tplRun(t, j, "varsa.html", data1, "<title>Hello World A</title>", "Key=Value A")
		tplRun(t, j, "sub2a.html", data1, "<MAIN>", "<SUB1>", "<SUB2A>", "</SUB2A>", "</SUB1>", "</MAIN>")
		tplRun(t, j, "sub1a.html", data1, "<MAIN>", "<SUB1>", "</SUB1>", "<Plain>", "</MAIN>")
	}
}

var tplMain string = `<MAIN>
{{ template "body" }}
{{ include "plaintext.html" }}
</MAIN>`

var tplSub1 string = `{{ extends "Folder/Main-file_name.html.dji" }}
{{ define "body" }}
<SUB1>

{{ template "content" .Data }}
</SUB1>
{{ end }}

{{ define "content" }}
	{{ if eq 1 1 }}
		{{ range .Slice }}
			{{ . }}
	{{ else }}
		NO SLICE
		{{ end }}
	{{ else }}
	{{ end }}
<DEFAULT CONTENT>
{{ end }}
`

var tplSub2 string = `
{{ extends 'sub1.html' }}

{{ define "content" }}
<SUB2></SUB2>
{{ end }}`

var tplSub2a string = `
{{ extends 'sub1.html' }}
{{ define "content" }}
<SUB2A></SUB2A>
{{ end }}`

func tmplfnc() error {
	return nil
}

func TestConfigure(t *testing.T) {
	m := make(map[string]string)
	m["testingTmpl"] = "testing template"
	mm := make(map[string]interface{})
	mm["testingFn"] = tmplfnc
	c := New(CacheOn(TLRUCache(1)), Loaders(MapLoader(m)), TemplateFunctions(mm))
	c.Configure()
	if c.Cache == nil || c.cached == false {
		t.Errorf("Cache configuration not configured as expected.")
	}
	_, err := c.Fetch("testingTmpl")
	if err != nil {
		t.Errorf("Loaders Conf function not configured as expected.")
	}
	if _, ok := c.GetFuncs()["testingFn"]; !ok {
		t.Errorf("TemplateFunctions not configured as expected.")
	}
}

var getTests = []struct {
	name       string
	keyToAdd   string
	keyToGet   string
	expectedOk bool
}{
	{"hit", "testing/template/1", "testing/template/1", true},
	{"miss", "testing/template/1", "testing/template/missing", false},
}

func TestGet(t *testing.T) {
	t1, _ := template.New("testing/template/1").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
	for _, tt := range getTests {
		lru := TLRUCache(0)
		lru.Add(tt.keyToAdd, t1)
		val, ok := lru.Get(tt.keyToGet)
		if ok != tt.expectedOk {
			t.Fatalf("%s: cache hit = %v; want %v", tt.name, ok, !ok)
		} else if ok && val != t1 {
			t.Fatalf("%s expected get to return template t1 but got %v", tt.name, val)
		}
	}
}

func TestRemove(t *testing.T) {
	t2, _ := template.New("testing/template/2").Parse(`{{define "T"}}Hello, {{.}}!{{end}}`)
	lru := TLRUCache(0)
	lru.Add("t2Key", t2)
	if val, ok := lru.Get("t2Key"); !ok {
		t.Fatal("TestRemove returned no match")
	} else if val != t2 {
		t.Fatalf("TestRemove failed. Expected %d, got %v", t2, val)
	}
	lru.Remove("t2Key")
	if _, ok := lru.Get("t2Key"); ok {
		t.Fatal("TestRemove returned a removed entry")
	}
}
