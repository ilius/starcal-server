package scal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type HMS struct {
	Hour   uint8
	Minute uint8
	Second uint8
}

type DHMS struct {
	HMS
	Days uint
}

type HMSRange struct {
	Start HMS
	End   HMS
}

func (hms HMS) String() string {
	return fmt.Sprintf("%.2d:%.2d:%.2d", hms.Hour, hms.Minute, hms.Second)
}

func (hms HMS) GetTotalSeconds() int {
	return int(hms.Hour)*3600 + int(hms.Minute)*60 + int(hms.Second)
}

func (hms HMS) GetFloatHour() float64 {
	return float64(hms.Hour) + float64(hms.Minute)/60.0 + float64(hms.Second)/3600.0
}

func (hms HMS) IsValid() bool {
	return hms.Hour >= 0 && hms.Hour < 24 &&
		hms.Minute >= 0 && hms.Minute < 60 &&
		hms.Second >= 0 && hms.Second < 60
}

func (dhms DHMS) IsValid() bool {
	return dhms.HMS.IsValid() && dhms.Days >= 0
}

func (dhms DHMS) String() string {
	return fmt.Sprintf("%d %s", dhms.Days, dhms.HMS.String())
}

func (hms HMSRange) IsValid() bool {
	return hms.Start.IsValid() && hms.End.IsValid()
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
	return HMS{uint8(h), uint8(m), uint8(s)}, nil
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
		Days: uint(days),
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
	hourInt := uint8(fh)
	hourPortion := fh - float64(hourInt)
	minuteFloat := hourPortion * 60.0
	minuteInt := uint8(minuteFloat)
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
	secondInt := uint8(secondFloat)
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
