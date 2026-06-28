package entity

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type Series struct {
	Id        int64     `db:"id"`
	ParentId  *int64    `db:"parent_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

type SeriesArguments struct {
	ParentId *int64
	Title    string
}

func (sa *SeriesArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"parent_id": sa.ParentId,
		"title":     sa.Title,
	}
}

func (s *Series) GetNumberingUrlString() string {
	lenSerial := len(strconv.Itoa(int(s.Id)))
	serialString := fmt.Sprintf("%d%d", lenSerial, int(s.Id))

	centuryString := strconv.Itoa((s.CreatedAt.Year()-1)/100 + 1)
	lenCentury := len(centuryString)
	centuryString = fmt.Sprintf("%d%s", lenCentury, centuryString)

	yearString := strconv.Itoa(s.CreatedAt.Year() % 100)
	lenYear := len(yearString)
	yearString = fmt.Sprintf("%d%s", lenCentury, yearString)

	dayString := strconv.Itoa(s.CreatedAt.YearDay())
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

func (s *Series) GetTimeRequiredString() string {
	return "-1"
}

func (s *Series) GetTitle() string {
	return s.Title
}
