package utils

import (
	"errors"
	//"fmt"
	"strconv"
	"strings"
)

const (
	u_second = "s"
	u_minute = "m"
	u_hour   = "h"
	u_day    = "d"
	u_week   = "w"
)

var durationUnits = map[string]int{
	u_second: 1,
	u_minute: 60,
	u_hour:   3600,
	u_day:    3600 * 24,
	u_week:   3600 * 24 * 7,
}

type Duration struct {
	Value       float64
	UnitString  string
	UnitSeconds int
}

func (dur Duration) IsValid() bool {
	// FIXME
	return dur.Value >= 0
}

func ParseDuration(str string) (Duration, error) {
	// format: "2 s", "2 m", "2 h", "2 d", "2 w"
	parts := strings.Split(str, " ")
	if len(parts) != 2 {
		return Duration{},
			errors.New("invalid Duration string '" + str + "'")
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Duration{}, err
	}
	unitString := parts[1]
	unitSeconds, ok := durationUnits[unitString]
	if !ok {
		return Duration{}, errors.New(
			"invalid Duration string '" + str + "', bad unit",
		)
	}
	return Duration{
		Value:       value,
		UnitString:  unitString,
		UnitSeconds: unitSeconds,
	}, nil
}
