package entity

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type Work struct {
	Id        int64     `db:"id"`
	SeriesId  *int64    `db:"series_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

type WorkArguments struct {
	SeriesId *int64
	Title    string
}

func (w *Work) GetNumberingUrlString() string {
	lenSerial := len(strconv.Itoa(int(w.Id)))
	serialString := fmt.Sprintf("%d%d", lenSerial, int(w.Id))

	centuryString := strconv.Itoa((w.CreatedAt.Year()-1)/100 + 1)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(w.CreatedAt.Year() % 100)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(w.CreatedAt.YearDay())
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

func (w *Work) GetTimeRequiredString() string {
	return "-1"
}

func (w *Work) GetTitle() string {
	return w.Title
}

func (w *Work) GetName(prefix string) string {
	return prefix + "work"
}

func (wa *WorkArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"series_id": wa.SeriesId,
		"title":     wa.Title,
	}
}
