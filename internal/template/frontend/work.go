package frontend

import (
	"WorksKeeper/internal/repository/entity"
	"fmt"
	"html/template"
	"strconv"
)

type TemplateWork struct {
	Work           *entity.Work
	TemplateCanvas Templatable
}

func (tw *TemplateWork) GetNumberingUrlString() string {
	lenSerial := len(strconv.Itoa(int(tw.Work.Id)))
	serialString := fmt.Sprintf("%d%d", lenSerial, int(tw.Work.Id))

	centuryString := strconv.Itoa((tw.Work.CreatedAt.Year()-1)/100 + 1)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(tw.Work.CreatedAt.Year() % 100)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(tw.Work.CreatedAt.YearDay())
	lenDay := len(dayString)
	dayString = fmt.Sprintf("%d%s", lenDay, dayString)

	if lenSerial > 6 || lenCentury > 3 || lenYear > 3 || lenDay > 3 {
		panic("invalid numbering provided when creating url")
	}

	return fmt.Sprintf("/work/%s%s%s%s",
		serialString,
		centuryString,
		yearString,
		dayString)
}

func (tw *TemplateWork) GetTimeRequiredString() string {
	return "-1"
}

func (tw *TemplateWork) GetTitle() string {
	return tw.Work.Title
}

func (tw *TemplateWork) GetName(prefix string) string {
	return prefix + "work"
}

func (tw *TemplateWork) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/listing/%s.html", fileName)
}

func (tw *TemplateWork) ToEditableHtml() template.HTML {
	return mustToEditableHtml(tw)
}

func (tw *TemplateWork) ToHtml() template.HTML {
	return mustToHtml(tw)
}
