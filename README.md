#Djinn
html templating tools in Go using standard library html/template. Templates are
used as html/template templates in syntax, and are actual \*template.Template instances

[![GoDoc](https://godoc.org/github.com/thrisp/djinn?status.png)](https://godoc.org/github.com/thrisp/djinn)
[![Build Status](https://travis-ci.org/thrisp/djinn.svg?branch=develop)](https://travis-ci.org/thrisp/djinn)
[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/thrisp/djinn/master/LICENSE)

[http://thrisp.github.io/djinn](http://thrisp.github.io/djinn)

install:

```go get github.com/thrisp/djinn```

quickstart:

```go
package main

import 'github.com/thrisp/djinn'

func main() {

    m := map[string]string{
		"hello.dji":  `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
    }

    J := djinn.New()
    
    J.AddLoaders(&MapLoader{m: &m})

    type TemplateData struct {
	    Title string
	    Data  map[string]interface{}
    } 

    data := &TemplateData{
		Title: "Hello World",
		Data: map[string]interface{}{
			"Key":   "Value",
		},
	}

    w := &bytes.Buffer{}

	err := J.Render(w, "hello.dji", data)
}
```
