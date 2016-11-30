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

type DHMS struct {
	HMS
	Days int
}

type HMSRange struct {
	Start HMS
	End   HMS
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

func (self HMS) IsValid() bool {
	return self.Hour >= 0 && self.Hour < 24 &&
		self.Minute >= 0 && self.Minute < 60 &&
		self.Second >= 0 && self.Second < 60
}
func (self DHMS) IsValid() bool {
	return self.HMS.IsValid() && self.Days >= 0
}
func (self HMSRange) IsValid() bool {
	return self.Start.IsValid() && self.End.IsValid()
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

func ParseDHMS(str string) (DHMS, error) {
	// Days and HMS, format: "365 23:55:55"
	parts := strings.Split(str, " ")
	if len(parts) != 2 {
		return DHMS{},
			errors.New("invalid DHMS string '" + str + "'")
	}
	days, err := strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return DHMS{}, err
	}
	hms, err := ParseHMS(parts[1])
	if err != nil {
		return DHMS{}, err
	}
	return DHMS{
		HMS:  hms,
		Days: int(days),
	}, nil
}

func ParseHMSRange(str string) (HMSRange, error) {
	// format: "14:30:00 15:30:00"
	parts := strings.Split(str, " ")
	if len(parts) != 2 {
		return HMSRange{},
			errors.New("invalid HMS Range string '" + str + "'")
	}
	start, err := ParseHMS(parts[0])
	if err != nil {
		return HMSRange{}, err
	}
	end, err := ParseHMS(parts[1])
	if err != nil {
		return HMSRange{}, err
	}
	return HMSRange{start, end}, nil
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
