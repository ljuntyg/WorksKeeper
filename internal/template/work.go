package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type WorkData struct {
	TemplateWork frontend.Templatable
}

func (wd *WorkData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/work-page.html")
	templ.Execute(rw, wd)
}
