package frontend

import (
	"WorksKeeper/internal/repository/entity"
)

type TemplateSeries struct {
	Series    *entity.Series
	Listables []Listable
}

func (ts *TemplateSeries) GetTitle() string {
	return ts.Series.Title
}

func (ts *TemplateSeries) GetNumberingUrlString() string {
	return ts.Series.GetNumberingUrlString()
}

func (ts *TemplateSeries) GetTimeRequiredString() string {
	return "-1"
}
