package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateCanvas struct {
	Canvas        *entity.Canvas
	TemplateGroup Templatable
}

func (tc *TemplateCanvas) GetName(prefix string) string {
	return prefix + "canvas"
}

func (tc *TemplateCanvas) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/container/%s.html", fileName)
}

func (tc *TemplateCanvas) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tc)
}

func (tc *TemplateCanvas) ToHtml() template.HTML {
	return mustToHtml(tc)
}
