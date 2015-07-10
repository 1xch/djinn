package djinn

import (
	"bytes"
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

var J1 *Djinn = New(Loaders(NewMapLoader(m1), NewMapLoader(m2)))

var J2 *Djinn = New(Loaders(NewDirLoader("./test/templates"), NewDirLoader("./test/additional/templates")))

var J3 *Djinn = New(Loaders(NewMapLoader(m1), NewDirLoader("./test/additional/templates")))

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
