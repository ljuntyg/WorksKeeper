package frontend

import (
	"WorksKeeper/internal/repository"
	"fmt"
	"html/template"
	"strconv"
)

type TemplateSeries struct {
	Series           *repository.Series
	TemplateListings []Templatable
}

func (ts *TemplateSeries) GetName(prefix string) string {
	return prefix + "series"
}

func (ts *TemplateSeries) GetPath(fileName string) string {
	return fmt.Sprintf("./resources/templates/listing/%s.html", fileName)
}

func (ts *TemplateSeries) ToEditableHtml() template.HTML {
	return mustToEditableHtml(ts)
}

func (ts *TemplateSeries) ToHtml() template.HTML {
	return mustToHtml(ts)
}

func (ts *TemplateSeries) GetTitle() string {
	return ts.Series.Title
}

func (ts *TemplateSeries) GetNumberingUrlString() string {
	lenSerial := len(strconv.Itoa(int(ts.Series.Id)))
	serialString := fmt.Sprintf("%d%d", lenSerial, int(ts.Series.Id))

	centuryString := strconv.Itoa((ts.Series.CreatedAt.Year()-1)/100 + 1)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(ts.Series.CreatedAt.Year() % 100)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(ts.Series.CreatedAt.YearDay())
	lenDay := len(dayString)
	dayString = fmt.Sprintf("%d%s", lenDay, dayString)

	if lenSerial > 6 || lenCentury > 3 || lenYear > 3 || lenDay > 3 {
		panic("invalid numbering provided when creating url")
	}

	return fmt.Sprintf("/series/%s%s%s%s",
		serialString,
		centuryString,
		yearString,
		dayString)
}

func (ts *TemplateSeries) GetTimeRequiredString() string {
	return "-1"
}
