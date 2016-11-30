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

func (self Date) String() string {
	return fmt.Sprintf("%.4d/%.2d/%.2d", self.Year, self.Month, self.Day)
}
func (self Date) Repr() string {
	return fmt.Sprintf("scal.Date{%d, %d, %d}", self.Year, self.Month, self.Day)
}
func (self Date) IsValid() bool {
	return self.Month > 0 && self.Month < 13 && self.Day > 0 && self.Day < 40
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

func (self DateHMS) String() string {
	return self.Date.String() + " " + self.HMS.String()
}
func (self DateHMS) Repr() string {
	return fmt.Sprintf("scal.DateHMS{{%s}, {%s}}", self.Date, self.HMS)
}
func (self DateHMS) IsValid() bool {
	return self.Date.IsValid() && self.HMS.IsValid()
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
