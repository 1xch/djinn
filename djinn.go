package djinn

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
)

type (
	Djinn struct {
		Loaders []TemplateLoader
		FuncMap map[string]interface{}
		cache   *TLRUCache
	}

	Node struct {
		Name string
		Src  string
	}
)

var (
	re_extends     *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
	re_defineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	re_templateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
	err            error
)

// A blank instance with a default cache
func New() *Djinn {
	d := &Djinn{}
	d.Loaders = make([]TemplateLoader, 0)
	d.FuncMap = make(map[string]interface{})
	d.cache = NewTLRUCache(50)
	return d
}

func (d *Djinn) AddLoaders(loaders ...TemplateLoader) {}

func (d *Djinn) Render(w io.Writer, name string, data interface{}) error {
	if tmpl, ok := d.cache.Get(name); ok {
		err = tmpl.Execute(w, data)
	} else {
		tmpl, err := d.assemble(name)
		if err != nil {
			return err
		}
		if tmpl == nil {
			return Errf("Nil template named %s", name)
		}
		err = tmpl.Execute(w, data)
	}

	if err != nil {
		return err
	}

	return nil
}

func (d *Djinn) FetchTemplate(w io.Writer, name string) (*template.Template, error) {
	if tmpl, ok := d.cache.Get(name); ok {
		return tmpl, nil
	} else {
		return d.assemble(name)
	}
}

func (d *Djinn) get(name string) (string, error) {
	for _, l := range d.Loaders {
		t, err := l.Load(name)
		if err == nil {
			return t, nil
		}
	}
	return "", Errf("Template %s does not exist", name)
}

func (d *Djinn) add(stack *[]*Node, name string) error {
	tplSrc, err := d.get(name)

	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return Errf("Empty Template named %s", name)
	}

	extendsMatches := re_extends.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := d.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = re_extends.ReplaceAllString(tplSrc, "")
	}

	node := &Node{
		Name: name,
		Src:  tplSrc,
	}

	*stack = append((*stack), node)

	return nil
}

func (d *Djinn) assemble(name string) (*template.Template, error) {
	stack := []*Node{}

	err := d.add(&stack, name)

	if err != nil {
		return nil, err
	}

	blocks := map[string]string{}
	blockId := 0

	var rootTemplate *template.Template

	for _, node := range stack {
		node.Src = re_defineTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := re_defineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	for i, node := range stack {
		node.Src = re_templateTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
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
			thisTemplate = template.New(node.Name)
			rootTemplate = thisTemplate
		} else {
			thisTemplate = rootTemplate.New(node.Name)
		}

		thisTemplate.Funcs(d.FuncMap)

		_, err := thisTemplate.Parse(node.Src)
		if err != nil {
			return nil, err
		}
	}

	d.cache.Add(name, rootTemplate)

	return rootTemplate, nil
}
