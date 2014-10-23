package djinn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type TemplateData struct {
	Title string
	Data  map[string]interface{}
}

func MustContain(t *testing.T, str *string, check string) {
	index := strings.Index(*str, check)
	if index < 0 {
		fmt.Printf("Template did not contain %s\n", check)
		t.Fail()
		return
	}
	*str = (*str)[index:]
}

func TplRun(t *testing.T, j *Djinn, name string, data interface{}, check ...string) {
	w := &bytes.Buffer{}
	err := j.Render(w, name, data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	//from cache
	j.Render(w, name, data)

	str := w.String()
	fmt.Println(str)

	for _, c := range check {
		MustContain(t, &str, c)
	}

}

func TestTemplate(t *testing.T) {
	m1 := map[string]string{
		"vars.html":                      `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
		"Folder/Main-file_name.html.dji": tplMain,
		"sub1.html":                      tplSub1,
		"sub2.html":                      tplSub2,
	}

	m2 := map[string]string{
		"varsa.html":                      `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
		"Folder/Main-file_namea.html.dji": tplMain,
		"sub1a.html":                      tplSub1,
		"sub2a.html":                      tplSub2a,
	}

	loader1 := NewMapLoader(m1)

	loader2 := NewMapLoader(m2)

	loader3 := NewDirLoader("./test/templates")

	loader4 := NewDirLoader("./test/additional/templates")

	j1 := New(Loaders(loader1, loader2))

	j2 := New(Loaders(loader3, loader4))

	j3 := New(Loaders(loader1, loader4))

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
		j1,
		j2,
		j3,
	}

	for _, j := range js {
		TplRun(t, j, "vars.html", data, "<title>Hello World</title>", "Key=Value")
		TplRun(t, j, "sub2.html", data, "<MAIN>", "<SUB1>", "<SUB2>", "</SUB2>", "</SUB1>", "</MAIN>")
		TplRun(t, j, "sub1.html", data, "<MAIN>", "<SUB1>", "</SUB1>", "</MAIN>")
		TplRun(t, j, "varsa.html", data1, "<title>Hello World A</title>", "Key=Value A")
		TplRun(t, j, "sub2a.html", data1, "<MAIN>", "<SUB1>", "<SUB2A>", "</SUB2A>", "</SUB1>", "</MAIN>")
		TplRun(t, j, "sub1a.html", data1, "<MAIN>", "<SUB1>", "</SUB1>", "</MAIN>")
	}
}

var tplMain string = `<MAIN>
{{ template "body" }}
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
