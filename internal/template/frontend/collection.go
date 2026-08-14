package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateCollection struct {
	Collection     *repository.Collection
	TemplateSeries Templatable
}

func (tc *TemplateCollection) GetName(prefix string) string {
	return prefix + "collection"
}

func (tc *TemplateCollection) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/container/%s.html", fileName)
}

func (tc *TemplateCollection) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tc)
}

func (tc *TemplateCollection) ToHtml() template.HTML {
	return mustToHtml(tc)
}
