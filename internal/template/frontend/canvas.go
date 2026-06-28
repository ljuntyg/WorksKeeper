package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateCanvas struct {
	Canvas         *entity.Canvas
	TemplateGroups []Templatable
}

func (tc *TemplateCanvas) GetName(prefix string) string {
	return tc.Canvas.GetName(prefix)
}

func (tc *TemplateCanvas) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/%s.html", fileName)
}

func (tc *TemplateCanvas) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tc)
}

func (tc *TemplateCanvas) ToHtml() template.HTML {
	return mustToHtml(tc)
}
