package jingo

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type TemplateData struct {
	Title string
	//Path  string
	//User  interface{}
	//Nav   map[string]string
	Data map[string]interface{}
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

func TplRun(t *testing.T, j *Jingo, name string, data interface{}, check ...string) {
	w := &bytes.Buffer{}
	err := j.Render(w, name, data)
	if err != nil {
		fmt.Println(err)
		t.Fail()
		return
	}

	str := w.String()
	fmt.Println(str)

	for _, c := range check {
		MustContain(t, &str, c)
	}
}

func TestTemplate(t *testing.T) {
	j1 := &Jingo{
		Loader: &MapLoader{
			"vars.html":                       `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
			"Folder/Main-file_name.html.twig": tplMain,
			"sub1.html":                       tplSub1,
			"sub2.html":                       tplSub2,
		},
	}

	j2 := &Jingo{
		Loader: &DirLoader{
			BasePath: "./test/templates",
		},
	}

	data := &TemplateData{
		Title: "Hello World",
		Data: map[string]interface{}{
			"Key":   "Value",
			"Slice": []string{"One", "Two", "Three"},
		},
	}

	jingoes := []*Jingo{
		j1,
		j2,
	}

	for _, j := range jingoes {
		TplRun(t, j, "vars.html", data, "<title>Hello World</title>", "Key=Value")
		TplRun(t, j, "sub2.html", data, "<MAIN>", "<SUB1>", "<SUB2>", "</SUB2>", "</SUB1>", "</MAIN>")
		TplRun(t, j, "sub1.html", data, "<MAIN>", "<SUB1>", "</SUB1>", "</MAIN>")
	}
}

var tplMain string = `<MAIN>
{{ template "body" }}
</MAIN>`

var tplSub1 string = `{{ extends "Folder/Main-file_name.html.twig" }}
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
