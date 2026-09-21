package template

import (
	"WorksKeeper/internal/template/frontend"
	"net/http"
)

type EditSeriesData struct {
	TemplateSeries *frontend.TemplateSeries
	IsEditing      bool
}

func (esd *EditSeriesData) ExecuteTemplate(rw http.ResponseWriter) {
	templ := MustLoadTemplate("./resources/templates/edit-series-page.html")
	templ.Execute(rw, esd)
}
