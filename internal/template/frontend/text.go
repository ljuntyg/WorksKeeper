package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateText struct {
	Text *repository.Text
}

func (tt *TemplateText) GetName(prefix string) string {
	return prefix + "text"
}

func (tt *TemplateText) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tt *TemplateText) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tt)
}

func (tt *TemplateText) ToHtml() template.HTML {
	return mustToHtml(tt)
}
