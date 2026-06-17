package logic

import (
	"github.com/agoXQ/QuantLab/app/market/domain/valueobject"
)

// parseDateRange converts wire-format date strings into a domain DateRange.
//
// Empty strings result in a zero value, which the application service treats
// as "no bound". This keeps logic files free of repetitive parsing code.
func parseDateRange(start, end string) (valueobject.DateRange, error) {
	var rng valueobject.DateRange
	if start != "" {
		t, err := valueobject.ParseDate(start)
		if err != nil {
			return rng, err
		}
		rng.Start = t
	}
	if end != "" {
		t, err := valueobject.ParseDate(end)
		if err != nil {
			return rng, err
		}
		rng.End = t
	}
	return rng, nil
}
