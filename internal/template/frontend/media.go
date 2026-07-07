package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateMedia struct {
	Media           *entity.Media
	Sources         []entity.Source
	TemplateCaption Templatable
}

func (tm *TemplateMedia) GetName(prefix string) string {
	return prefix + "media"
}

func (tm *TemplateMedia) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tm *TemplateMedia) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tm)
}

func (tm *TemplateMedia) ToHtml() template.HTML {
	return mustToHtml(tm)
}
