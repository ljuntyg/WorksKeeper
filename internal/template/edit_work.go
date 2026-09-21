package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type EditWorkData struct {
	TemplateWork frontend.Templatable
	IsEditing    bool
}

func (ewd *EditWorkData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/edit-work-page.html")
	templ.Execute(rw, ewd)
}
