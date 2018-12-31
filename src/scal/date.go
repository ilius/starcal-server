package scal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Date struct {
	Year  int
	Month int
	Day   int
}

func (date Date) String() string {
	return fmt.Sprintf("%.4d/%.2d/%.2d", date.Year, date.Month, date.Day)
}
func (date Date) Repr() string {
	return fmt.Sprintf("scal.Date{%d, %d, %d}", date.Year, date.Month, date.Day)
}
func (date Date) IsValid() bool {
	return date.Month > 0 && date.Month < 13 && date.Day > 0 && date.Day < 40
}

func ParseDate(str string) (Date, error) {
	parts := strings.Split(str, "/")
	if len(parts) != 3 {
		return Date{},
			errors.New("invalid Date string '" + str + "'")
	}
	var err error
	var y, m, d int64
	y, err = strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return Date{}, err
	}
	m, err = strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return Date{}, err
	}
	d, err = strconv.ParseInt(parts[2], 10, 0)
	if err != nil {
		return Date{}, err
	}
	return Date{int(y), int(m), int(d)}, nil

}

func ParseDateList(str string) ([]Date, error) {
	parts := strings.Split(str, " ")
	dates := make([]Date, len(parts))
	for index, part := range parts {
		date, err := ParseDate(part)
		if err != nil {
			return []Date{}, err
		}
		dates[index] = date
	}
	return dates, nil
}

type DateHMS struct {
	Date
	HMS
}

func (dt DateHMS) String() string {
	return dt.Date.String() + " " + dt.HMS.String()
}
func (dt DateHMS) Repr() string {
	return fmt.Sprintf("scal.DateHMS{{%s}, {%s}}", dt.Date, dt.HMS)
}
func (dt DateHMS) IsValid() bool {
	return dt.Date.IsValid() && dt.HMS.IsValid()
}
func ParseDateHMS(str string) (DateHMS, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 2 {
		return DateHMS{},
			errors.New("invalid DateHMS string '" + str + "'")
	}
	date, err := ParseDate(parts[0])
	if err != nil {
		return DateHMS{}, err
	}
	hms, err := ParseHMS(parts[1])
	if err != nil {
		return DateHMS{}, err
	}
	return DateHMS{
		Date: date,
		HMS:  hms,
	}, nil
}
