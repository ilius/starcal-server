package hijri

import "fmt"

import "scal"

func PrintIsLeap(start_year int, end_year int) {
	for year := start_year; year < end_year; year++ {
		fmt.Printf(
			"        %v: %v,\n",
			year,
			IsLeap(year),
		)
	}
}

func PrintToJd(start_year int, end_year int) {
	var date scal.Date
	var jd int
	for year := start_year; year < end_year; year++ {
		for month := 1; month <= 12; month++ {
			date = scal.Date{year, month, 1}
			jd = ToJd(date)
			fmt.Printf(
				"        %v: %v,\n",
				date.Repr(),
				jd,
			)
		}
	}
}
