package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type HomeData struct {
	TemplateCollection frontend.Templatable
	HasMore            bool
}

func (hd *HomeData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/home-page.html")
	templ.Execute(rw, hd)
}
