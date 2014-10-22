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
		Cache
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

func Empty() *Djinn {
	return &Djinn{}
}

// A blank instance with a default Loaders & FuncMap maps made but empty,
// and a TLRUCache.
func New() *Djinn {
	j := Empty()
	j.Loaders = make([]TemplateLoader, 0)
	j.FuncMap = make(map[string]interface{})
	j.Cache = NewTLRUCache(50)
	return j
}

func (j *Djinn) AddLoaders(loaders ...TemplateLoader) {
	for _, l := range loaders {
		j.Loaders = append(j.Loaders, l)
	}
	return
}

func (j *Djinn) Render(w io.Writer, name string, data interface{}) error {
	if tmpl, ok := j.Cache.Get(name); ok {
		err = tmpl.Execute(w, data)
	} else {
		tmpl, err := j.assemble(name)
		if err != nil {
			return err
		}
		if tmpl == nil {
			return DjinnError("Nil template named %s", name)
		}
		err = tmpl.Execute(w, data)
	}

	if err != nil {
		return err
	}

	return nil
}

func (j *Djinn) FetchTemplate(w io.Writer, name string) (*template.Template, error) {
	if tmpl, ok := j.Cache.Get(name); ok {
		return tmpl, nil
	} else {
		return j.assemble(name)
	}
}

func (j *Djinn) getTemplate(name string) (string, error) {
	for _, l := range j.Loaders {
		t, err := l.Load(name)
		if err == nil {
			return t, nil
		}
	}
	return "", DjinnError("Template %s does not exist", name)
}

func (j *Djinn) add(stack *[]*Node, name string) error {
	tplSrc, err := j.getTemplate(name)

	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return DjinnError("Empty Template named %s", name)
	}

	extendsMatches := re_extends.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := j.add(stack, extendsMatches[1])
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

func (j *Djinn) assemble(name string) (*template.Template, error) {
	stack := []*Node{}

	err := j.add(&stack, name)

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

		thisTemplate.Funcs(j.FuncMap)

		_, err := thisTemplate.Parse(node.Src)
		if err != nil {
			return nil, err
		}
	}

	j.Cache.Add(name, rootTemplate)

	return rootTemplate, nil
}
