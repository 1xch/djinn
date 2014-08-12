#Jingo
html templating in Go

[![GoDoc](https://godoc.org/github.com/thrisp/jingo?status.png)](https://godoc.org/github.com/thrisp/jingo)
[![Build Status](https://travis-ci.org/thrisp/jingo.svg?branch=develop)](https://travis-ci.org/thrisp/jingo)

install:

go get github.com/thrisp/jingo

quickstart:

```go
package main

import 'github.com/thrisp/jingo'

func main() {

    m := map[string]string{
		"hello.jingo":  `<title>{{.Title}}</title> Key={{ .Data.Key }}`,
    }

    J := NewJingo()
    
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

	err := J.Render(w, "hello.jingo", data)
}
```
