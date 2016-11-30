package scal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type HMS struct {
	Hour   int
	Minute int
	Second int
}

func (self HMS) String() string {
	return fmt.Sprintf("%.2d:%.2d:%.2d", self.Hour, self.Minute, self.Second)
}
func (self HMS) GetTotalSeconds() int {
	return self.Hour*3600 + self.Minute*60 + self.Second
}
func (self HMS) GetFloatHour() float64 {
	return float64(self.Hour) + float64(self.Minute)/60.0 + float64(self.Second)/3600.0
}

func ParseHMS(str string) (HMS, error) {
	parts := strings.Split(str, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return HMS{},
			errors.New("invalid HMS string '" + str + "'")
	}
	h, h_err := strconv.ParseInt(parts[0], 10, 0)
	if h_err != nil {
		return HMS{}, h_err
	}
	m, m_err := strconv.ParseInt(parts[1], 10, 0)
	if m_err != nil {
		return HMS{}, m_err
	}
	var s int64
	var s_err error
	if len(parts) == 3 {
		s, s_err = strconv.ParseInt(parts[2], 10, 0)
		if s_err != nil {
			return HMS{}, s_err
		}
	} else {
		s = 0
	}
	return HMS{int(h), int(m), int(s)}, nil
}

func FloatHourToHMS(fh float64) HMS {
	hourInt := int(fh)
	hourPortion := fh - float64(hourInt)
	minuteFloat := hourPortion * 60.0
	minuteInt := int(minuteFloat)
	minutePortion := minuteFloat - float64(minuteInt)
	if minutePortion > 0.98 {
		minutePortion = 0.0
		minuteInt++
		if minuteInt == 60 {
			minuteInt = 0
			hourInt++
		}
	}
	secondFloat := minutePortion * 60
	secondInt := int(secondFloat)
	secondPortion := secondFloat - float64(secondInt)
	if secondPortion > 0.98 {
		//secondPortion = 0.0
		secondInt++
		if secondInt == 60 {
			secondInt = 0
			minuteInt++
			if minuteInt == 60 {
				minuteInt = 0
				hourInt++
			}
		}
	}
	return HMS{hourInt, minuteInt, secondInt}
}
