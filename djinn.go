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
		*conf
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

// Empty returns an empty Djinn with no configuration.
func Empty() *Djinn {
	return &Djinn{
		Loaders: make([]TemplateLoader, 0),
		FuncMap: make(map[string]interface{}),
	}
}

// New provides a Djinn with default configuration.
func New(opts ...Conf) *Djinn {
	j := Empty()
	j.conf = defaultconf()
	opts = append(opts, CacheOn(NewTLRUCache(50)))
	j.SetConf(opts...)
	return j
}

// Render excutes the template specified by name, with the supplied writer and
// data. Template is searched for in the cache, if enabled, then from assembling
// the from the Djinn loaders. Returns any errors ocurring during these steps.
func (j *Djinn) Render(w io.Writer, name string, data interface{}) error {
	if j.cacheon {
		if tmpl, ok := j.Cache.Get(name); ok {
			err = tmpl.Execute(w, data)
			return nil
		}
	}

	tmpl, err := j.assemble(name)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return DjinnError("Nil template named %s", name)
	}
	err = tmpl.Execute(w, data)

	if err != nil {
		return err
	}

	return nil
}

// It's happening. Given a string name, Fetch attempts to get a
// *template.Template or returns an error.
func (j *Djinn) Fetch(name string) (*template.Template, error) {
	if j.cacheon {
		if tmpl, ok := j.Cache.Get(name); ok {
			return tmpl, nil
		}
	}
	return j.assemble(name)
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

	if j.cacheon {
		j.Cache.Add(name, rootTemplate)
	}

	return rootTemplate, nil
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

func (j *Djinn) getTemplate(name string) (string, error) {
	for _, l := range j.Loaders {
		t, err := l.Load(name)
		if err == nil {
			return t, nil
		}
	}
	return "", DjinnError("Template %s does not exist", name)
}
