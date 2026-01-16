package repository

import (
	"github.com/go-viper/mapstructure/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

func ParseToPgUUID(input string) (pgtype.UUID, error) {
	pgUUID := pgtype.UUID{}
	err := pgUUID.Scan(input)
	return pgUUID, err
}

func ParseToPgText(input any) (pgtype.Text, error) {
	pgText := pgtype.Text{}

	inputString, err := decode[string](input)
	if err != nil {
		return pgText, err
	}

	if inputString == "" {
		return pgText, nil
	}

	pgText.String = inputString
	pgText.Valid = true
	return pgText, nil
}

func ParseToPgDate(input any) (pgtype.Date, error) {
	pgDate := pgtype.Date{}

	var (
		date time.Time
		err  error
	)
	switch inp := input.(type) {
	case string:
		if inp == "" {
			return pgDate, nil
		}

		date, err = time.Parse(time.RFC3339, inp)
		if err != nil {
			return pgDate, err
		}
	case time.Time:
		if inp.IsZero() {
			return pgDate, nil
		}
		date = inp
	case int:
		if inp == 0 {
			return pgDate, nil
		}
		date = time.Unix(int64(inp), 0)
	case int64:
		if inp == 0 {
			return pgDate, nil
		}
		date = time.Unix(inp, 0)
	default:
		date, err = decode[time.Time](input)
		if err != nil {
			return pgDate, err
		}
	}

	pgDate.Time = date
	pgDate.Valid = true
	return pgDate, nil
}

func decode[T any](input any) (T, error) {
	var result T

	result, ok := input.(T)
	if ok {
		return result, nil
	}

	cfg := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
	}

	decoder, err := mapstructure.NewDecoder(&cfg)
	if err != nil {
		return result, err
	}
	err = decoder.Decode(input)
	return result, err
}
