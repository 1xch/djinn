#Jingo
html templating in Go

quickstart:

go get github.com/thrisp/jingo

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
