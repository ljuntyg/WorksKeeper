package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type HomeData struct {
	Listables []frontend.Listable
	HasMore   bool
}

func (hd *HomeData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/home-page.html")
	templ.Execute(rw, hd)
}
