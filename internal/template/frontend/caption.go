package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateCaption struct {
	Caption *entity.Caption
}

func (tc *TemplateCaption) GetName(prefix string) string {
	return prefix + "caption"
}

func (tc *TemplateCaption) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/content/%s.html", fileName)
}

func (tc *TemplateCaption) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tc)
}

func (tc *TemplateCaption) ToHtml() template.HTML {
	return mustToHtml(tc)
}
