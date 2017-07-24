package djinn

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
)

// The primary renderer, containing loaders, template functions, cache, & configuration.
type Djinn struct {
	Configuration
	*LoaderSet
	*FuncSet
	Cache
}

// Empty returns an empty Djinn with provided Config applied immediately.
func Empty(cnf ...Config) *Djinn {
	d := &Djinn{
		LoaderSet: NewLoaderSet(),
		FuncSet:   NewFuncSet(),
	}
	configure(d, cnf...)
	d.Configuration = newConfiguration(d)
	return d
}

// New provides a Djinn with default configuration.
func New(cnf ...Config) *Djinn {
	d := Empty()
	d.AddConfig(cnf...)
	return d
}

// Render excutes the template specified by name, with the supplied writer and
// data. Template is searched for in the cache, if enabled, then from assembling
// the from the tempalte Djinn loaders. Returns any ocurring errors.
func (d *Djinn) Render(w io.Writer, name string, data interface{}) error {
	if d.On() {
		if tmpl, ok := d.Cache.Get(name); ok {
			return tmpl.Execute(w, data)
		}
	}

	tmpl, err := d.assemble(name)

	if err != nil {
		return err
	}

	if tmpl == nil {
		return NilTemplateError(name)
	}

	return tmpl.Execute(w, data)
}

// Given a string name, Fetch attempts to get a *template.Template from cache
// or loaders, returning any error.
func (d *Djinn) Fetch(name string) (*template.Template, error) {
	if d.On() {
		if tmpl, ok := d.Cache.Get(name); ok {
			return tmpl, nil
		}
	}
	return d.assemble(name)
}

var (
	reExtendsTag  *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
	reIncludeTag  *regexp.Regexp = regexp.MustCompile(`{{ include ["']?([^"]*)["']? }}`)
	reDefineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	reTemplateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
)

func (d *Djinn) assemble(name string) (*template.Template, error) {
	stack := []*Node{}

	err := d.add(&stack, name)

	if err != nil {
		return nil, err
	}

	blocks := map[string]string{}
	blockId := 0

	for _, node := range stack {
		var errInReplace error = nil
		node.Src = reIncludeTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reIncludeTag.FindStringSubmatch(raw)
			templatePath := parsed[1]
			subTpl, err := d.getTemplate(templatePath)
			if err != nil {
				errInReplace = err
				return "[error]"
			}
			return subTpl
		})
		if errInReplace != nil {
			return nil, errInReplace
		}
	}

	for _, node := range stack {
		node.Src = reDefineTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reDefineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	var rootTemplate *template.Template

	for i, node := range stack {
		node.Src = reTemplateTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := reTemplateTag.FindStringSubmatch(raw)
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
			thisTemplate = template.New(node.Name)
			rootTemplate = thisTemplate
		} else {
			thisTemplate = rootTemplate.New(node.Name)
		}

		thisTemplate.Funcs(d.GetFuncs())

		_, err := thisTemplate.Parse(node.Src)
		if err != nil {
			return nil, err
		}
	}

	if d.On() {
		d.Cache.Add(name, rootTemplate)
	}

	return rootTemplate, nil
}

func (d *Djinn) getTemplate(name string) (string, error) {
	for _, l := range d.GetLoaders() {
		t, err := l.Load(name)
		if err == nil {
			return t, nil
		}
	}
	return "", NoTemplateError(name)
}

type Node struct {
	Name string
	Src  string
}

func (d *Djinn) add(stack *[]*Node, name string) error {
	tplSrc, err := d.getTemplate(name)
	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return EmptyTemplateError(name)
	}

	extendsMatches := reExtendsTag.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := d.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = reExtendsTag.ReplaceAllString(tplSrc, "")
	}

	node := &Node{
		Name: name,
		Src:  tplSrc,
	}

	*stack = append((*stack), node)

	return nil
}
