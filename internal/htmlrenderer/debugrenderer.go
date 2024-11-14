package htmlrenderer

import (
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/gin-gonic/gin/render"
)

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
