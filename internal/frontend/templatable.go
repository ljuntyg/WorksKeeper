package frontend

import "html/template"

type Templatable interface {
	/*Viewable
	Editable*/
}

type Viewable interface {
	Pathable
	ToHtml() template.HTML
}

type Editable interface {
	Pathable
	ToEditableHtml() template.HTML
}

type Pathable interface {
	GetName(prefix string) string
	GetPath(fileName string) string
}
