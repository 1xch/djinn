package jingo

import (
	"fmt"
	"html/template"
	"io"

	"regexp"
)

var re_extends *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
var re_defineTag *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
var re_templateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")

//macro

type Jingo struct {
	Loader  TemplateLoader
	FuncMap map[string]interface{} //template.FuncMap
}

type NamedTemplate struct {
	Name string
	Src  string
}

func (j *Jingo) Render(w io.Writer, name string, data interface{}) error {
	tpl, err := j.assemble(name)
	if err != nil {
		return err
	}

	if tpl == nil {
		return Errf("Nil template named %s", name)
	}

	err = tpl.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func (j *Jingo) GetTemplate(w io.Writer, name string) (*template.Template, error) {
	return j.assemble(name)
}

func (j *Jingo) add(stack *[]*NamedTemplate, name string) error {
	tplSrc, err := j.Loader.LoadTemplate(name)
	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return Errf("Empty Template named %s", name)
	}

	extendsMatches := re_extends.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 { //Did Match
		err := j.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = re_extends.ReplaceAllString(tplSrc, "")
	}
	namedTemplate := &NamedTemplate{
		Name: name,
		Src:  tplSrc,
	}
	*stack = append((*stack), namedTemplate)
	// The stack is ordered 'general' to 'specific'
	return nil
}

func (j *Jingo) assemble(name string) (*template.Template, error) {
	stack := []*NamedTemplate{}

	err := j.add(&stack, name)

	if err != nil {
		return nil, err
	}

	blocks := map[string]string{}
	blockId := 0

	var rootTemplate *template.Template
	for _, namedTemplate := range stack {
		namedTemplate.Src = re_defineTag.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_defineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	for i, namedTemplate := range stack {
		namedTemplate.Src = re_templateTag.ReplaceAllStringFunc(namedTemplate.Src, func(raw string) string {
			parsed := re_templateTag.FindStringSubmatch(raw)
			origName := parsed[1]
			replacedName, ok := blocks[origName]

			dot := "."
			if len(parsed) == 3 && len(parsed[2]) > 0 {
				dot = parsed[2]
			}
			if ok {
				return fmt.Sprintf(`{{ template "%s" %s }}`, replacedName, dot)
			} else {
				return ""
			}
		})

		var thisTemplate *template.Template

		if i == 0 {
			thisTemplate = template.New(namedTemplate.Name)
			rootTemplate = thisTemplate
		} else {
			thisTemplate = rootTemplate.New(namedTemplate.Name)
		}

		thisTemplate.Funcs(j.FuncMap)

		_, err := thisTemplate.Parse(namedTemplate.Src)
		if err != nil {
			return nil, err
		}
	}

	return rootTemplate, nil
}