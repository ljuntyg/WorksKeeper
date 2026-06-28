package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
)

type TemplateWork struct {
	Work           *entity.Work
	TemplateCanvas Templatable
}

func (tw *TemplateWork) GetName(prefix string) string {
	return tw.Work.GetName(prefix)
}

func (tw *TemplateWork) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/listable/%s.html", fileName)
}

func (tw *TemplateWork) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tw)
}

func (tw *TemplateWork) ToHtml() template.HTML {
	return mustToHtml(tw)
}
