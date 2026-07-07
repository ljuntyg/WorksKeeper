package entity

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Listing struct {
	Id             int64          `db:"id"`
	ParentSeriesId int64          `db:"parent_series_id"`
	Position       pgtype.Numeric `db:"position"`
	ListingType    string         `db:"listing_type"`
}

type ListingArguments struct {
	ParentSeriesId int64
	Position       pgtype.Numeric
	ListingType    string
}

func (la *ListingArguments) GetNamedArgs() pgx.NamedArgs {
	return pgx.NamedArgs{
		"parent_series_id": la.ParentSeriesId,
		"position":         la.Position,
		"listing_type":     la.ListingType,
	}
}
