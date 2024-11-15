// adapted from gin-contrib/multitemplate
// https://github.com/gin-contrib/multitemplate/blob/master/multitemplate.go

package htmlrenderer

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/gin-gonic/gin/render"
)

type Render struct {
	leftDelim  string
	rightDelim string
	funcs      template.FuncMap
	includes   *template.Template
	templates  map[string]*template.Template

	templateDir string
}

type Renderer interface {
	render.HTMLRender
	// Set the delimiters of the HTML renderer. Empty values will use the
	// default delimiters {{ and }}.
	Delims(left, right string)
	// Set the func map of the HTML renderer.
	Funcs(funcs template.FuncMap)
	// Set the directory for includes- templates that will be referenced from
	// all other templates.
	AddIncludes(dir string)
	// Set the directory for templates
	AddTemplates(dir string)
}

var (
	_ Renderer = &Render{}
)

func New() *Render {
	return &Render{
		includes:  template.New(""),
		templates: make(map[string]*template.Template),
	}
}

func (r *Render) Delims(left, right string) {
	r.leftDelim = left
	r.rightDelim = right
}

func (r *Render) Funcs(funcs template.FuncMap) {
	r.funcs = funcs
}

func (r *Render) AddIncludes(dir string) {
	r.includes = template.New("").
		Delims(r.leftDelim, r.rightDelim).
		Funcs(r.funcs)
	err := filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}

			r.includes, err = r.includes.ParseFiles(path)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}
}

func (r *Render) AddTemplates(dir string) {
	e := filepath.WalkDir(
		dir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			root, err := r.includes.Clone()
			if err != nil {
				return err
			}
			tmpl, err :=
				root.
					Delims(r.leftDelim, r.rightDelim).
					Funcs(r.funcs).
					ParseFiles(path)
			if err != nil {
				return err
			}
			r.templates[path] = tmpl
			fmt.Println(path)
			for _, t := range tmpl.Templates() {
				fmt.Println("-", t.Name())
			}
			return nil
		},
	)
	if e != nil {
		panic(e)
	}
	r.templateDir = dir
}

func (r *Render) Instance(name string, data any) render.Render {
	return render.HTML{
		Template: r.templates[filepath.Join(r.templateDir, name)],
		Name:     filepath.Base(name),
		Data:     data,
	}
}

// Variant of HTMLRenderer that allows for hot reloading. Every time a template is queried, it is
// reparsed and regenerated. For this reason, do not use this variant in production.
type DebugRender struct {
	leftDelim  string
	rightDelim string

	funcs template.FuncMap

	includeDir  string
	templateDir string
}

var (
	_ render.HTMLRender = &DebugRender{}
	_ Renderer          = &DebugRender{}
)

func NewDebug() *DebugRender {
	return &DebugRender{}
}

func (r *DebugRender) Delims(left, right string) {
	r.leftDelim = left
	r.rightDelim = right
}

func (r *DebugRender) Funcs(funcs template.FuncMap) {
	r.funcs = funcs
}

func (r *DebugRender) AddIncludes(dir string) {
	r.includeDir = dir
}

func (r *DebugRender) AddTemplates(dir string) {
	r.templateDir = dir
}

func (r *DebugRender) Instance(name string, data any) render.Render {
	root := template.New("").Delims(r.leftDelim, r.rightDelim).Funcs(r.funcs)
	err := filepath.WalkDir(
		r.includeDir,
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}

			root, err = root.ParseFiles(path)
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	tmpl := template.Must(root.ParseFiles(filepath.Join(r.templateDir, name)))
	return render.HTML{
		Template: tmpl,
		Name:     filepath.Base(name),
		Data:     data,
	}
}
