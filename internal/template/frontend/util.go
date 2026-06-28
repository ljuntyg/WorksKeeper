package frontend

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
)

func MustLoadTemplate(filePath string) *template.Template {
	tmpl, err := template.ParseFiles(filePath)
	if err != nil {
		log.Fatal(err)
	}

	return tmpl
}

func mustToEditableHtml(editable Editable) template.HTML {
	templ := MustLoadTemplate(editable.GetPath(editable.GetName("editable/")))

	var buf bytes.Buffer
	if err := templ.Execute(&buf, editable); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", editable.GetName("")))
	}

	return template.HTML(buf.String())
}

func mustToHtml(viewable Viewable) template.HTML {
	templ := MustLoadTemplate(viewable.GetPath(viewable.GetName("")))

	var buf bytes.Buffer
	if err := templ.Execute(&buf, viewable); err != nil {
		panic(fmt.Sprintf("unexpected error executing %s HTML template", viewable.GetName("")))
	}

	return template.HTML(buf.String())
}
