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
