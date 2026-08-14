package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
)

type TemplateContent struct {
	Content                    *repository.Content
	TemplateGroupOrTextOrMedia Templatable
}

func (tc *TemplateContent) GetName(prefix string) string {
	return prefix + "content"
}

func (tc *TemplateContent) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tc *TemplateContent) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tc)
}

func (tc *TemplateContent) ToHtml() template.HTML {
	return mustToHtml(tc)
}
