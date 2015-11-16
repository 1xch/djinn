package djinn

import (
	"fmt"
	"html/template"
	"io"
	"regexp"
)

// The primary srtuct for templates and rendering, containing loaders,
// template functions, cache, & configuration.
type Djinn struct {
	Loaders []TemplateLoader
	FuncMap map[string]interface{}
	Cache
	*conf
}

// Empty returns an empty Djinn with default configuration.
func Empty() *Djinn {
	return &Djinn{
		conf:    defaultconf(),
		Loaders: make([]TemplateLoader, 0),
		FuncMap: make(map[string]interface{}),
	}
}

// New provides a Djinn with default configuration & cache set to on.
func New(opts ...Conf) *Djinn {
	j := Empty()
	opts = append(opts, CacheOn(NewTLRUCache(100)))
	err := j.SetConf(opts...)
	if err != nil {
		panic(ConfigurationError(err.Error()))
	}
	return j
}

var NilTemplateError = Drror("nil template named %s").Out

// Render excutes the template specified by name, with the supplied writer and
// data. Template is searched for in the cache, if enabled, then from assembling
// the from the tempalte Djinn loaders. Returns any ocurring errors.
func (j *Djinn) Render(w io.Writer, name string, data interface{}) error {
	if j.CacheOn {
		if tmpl, ok := j.Cache.Get(name); ok {
			return tmpl.Execute(w, data)
		}
	}

	tmpl, err := j.assemble(name)

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
func (j *Djinn) Fetch(name string) (*template.Template, error) {
	if j.CacheOn {
		if tmpl, ok := j.Cache.Get(name); ok {
			return tmpl, nil
		}
	}
	return j.assemble(name)
}

type Node struct {
	Name string
	Src  string
}

var (
	re_extendsTag  *regexp.Regexp = regexp.MustCompile("{{ extends [\"']?([^'\"}']*)[\"']? }}")
	re_includeTag  *regexp.Regexp = regexp.MustCompile(`{{ include ["']?([^"]*)["']? }}`)
	re_defineTag   *regexp.Regexp = regexp.MustCompile("{{ ?define \"([^\"]*)\" ?\"?([a-zA-Z0-9]*)?\"? ?}}")
	re_templateTag *regexp.Regexp = regexp.MustCompile("{{ ?template \"([^\"]*)\" ?([^ ]*)? ?}}")
)

func (j *Djinn) assemble(name string) (*template.Template, error) {
	stack := []*Node{}

	err := j.add(&stack, name)

	if err != nil {
		return nil, err
	}

	blocks := map[string]string{}
	blockId := 0

	for _, node := range stack {
		var errInReplace error = nil
		node.Src = re_includeTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := re_includeTag.FindStringSubmatch(raw)
			templatePath := parsed[1]
			subTpl, err := j.getTemplate(templatePath)
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
		node.Src = re_defineTag.ReplaceAllStringFunc(node.Src, func(raw string) string {
			parsed := re_defineTag.FindStringSubmatch(raw)
			blockName := fmt.Sprintf("BLOCK_%d", blockId)
			blocks[parsed[1]] = blockName

			blockId += 1
			return "{{ define \"" + blockName + "\" }}"
		})
	}

	var rootTemplate *template.Template

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

	if j.CacheOn {
		j.Cache.Add(name, rootTemplate)
	}

	return rootTemplate, nil
}

var EmptyTemplateError = Drror("empty template named %s").Out

func (j *Djinn) add(stack *[]*Node, name string) error {
	tplSrc, err := j.getTemplate(name)

	if err != nil {
		return err
	}

	if len(tplSrc) < 1 {
		return EmptyTemplateError(name)
	}

	extendsMatches := re_extendsTag.FindStringSubmatch(tplSrc)
	if len(extendsMatches) == 2 {
		err := j.add(stack, extendsMatches[1])
		if err != nil {
			return err
		}
		tplSrc = re_extendsTag.ReplaceAllString(tplSrc, "")
	}

	node := &Node{
		Name: name,
		Src:  tplSrc,
	}

	*stack = append((*stack), node)

	return nil
}

var NoTemplateError = Drror("no template named %s").Out

func (j *Djinn) getTemplate(name string) (string, error) {
	for _, l := range j.Loaders {
		t, err := l.Load(name)
		if err == nil {
			return t, nil
		}
	}
	return "", NoTemplateError(name)
}
