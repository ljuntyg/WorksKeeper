package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type CanvasData struct {
	TemplateWork frontend.Templatable
	IsEditing    bool
}

func (cd *CanvasData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/canvas-page.html")
	templ.Execute(rw, cd)
}
