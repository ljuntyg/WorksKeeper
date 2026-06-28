package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type SeriesData struct {
	TemplateSeries frontend.Listable
}

func (sd *SeriesData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/series-page.html")
	templ.Execute(rw, sd)
}
