package indian_national

import "fmt"

import "scal"

func PrintIsLeap(startYear int, endYear int) {
	for year := startYear; year < endYear; year++ {
		fmt.Printf(
			"        %v: %v,\n",
			year,
			IsLeap(year),
		)
	}
}

func PrintToJd(startYear int, endYear int) {
	var date scal.Date
	var jd int
	for year := startYear; year < endYear; year++ {
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
