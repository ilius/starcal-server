// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
//
// Using kdelibs-4.4.0/kdecore/date/kcalendarsystemindiannational.cpp
//        Copyright (C) 2009 John Layt <john@layt.net>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License along
// with this program. If not, see <https://www.gnu.org/licenses/agpl.txt>.

package indian_national

import "scal"
import "scal/cal_types"
import "scal/cal_types/gregorian"

// ###### Common Globals #######

const (
	Name  = "indian_national"
	Desc  = "Indian National"
	Epoch = 1749994

	MinMonthLen uint8 = 30
	MaxMonthLen uint8 = 31

	AvgYearLen = 365.2425 // FIXME
)

var MonthNames = []string{
	"Chaitra", "Vaishākh", "Jyaishtha", "Āshādha", "Shrāvana", "Bhādrapad",
	"Āshwin", "Kārtik", "Agrahayana", "Paush", "Māgh", "Phālgun",
}
var MonthNamesAb = []string{
	"Cha", "Vai", "Jya", "Āsh", "Shr", "Bhā",
	"Āsw", "Kār", "Agr", "Pau", "Māg", "Phā",
}

// #############################

func init() {
	cal_types.RegisterCalType(
		Name,
		Desc,
		Epoch,
		MinMonthLen,
		MaxMonthLen,
		AvgYearLen,
		MonthNames,
		MonthNamesAb,
		IsLeap,
		ToJd,
		JdTo,
		GetMonthLen,
	)
}

func IsLeap(year int) bool {
	return gregorian.IsLeap(year + 78)
}

func ToJd(date scal.Date) int {
	// The calendar is closely synchronized to the Gregorian Calendar, always starting on the same day
	// We can use this and the regular sequence of days in months to do a simple conversion by finding
	// the Julian Day number of the first day of the year and adding on the required number of months
	// and days to get the final Julian Day number

	// Calculate the jd of 1 Chaitra for this year and how many days are in Chaitra this year
	// If a Leap Year, then 1 Chaitra == 21 March of the Gregorian year and Chaitra has 31 days
	// If not a Leap Year, then 1 Chaitra == 22 March of the Gregorian year and Chaitra has 30 days
	// Need to use dateToJulianDay() to calculate instead of setDate() to avoid the year 9999 validation
	var jdFirstDayOfYear int
	var daysInMonth1 int
	if IsLeap(date.Year) {
		jdFirstDayOfYear = gregorian.ToJd(scal.Date{date.Year + 78, 3, 21})
		daysInMonth1 = 31
	} else {
		jdFirstDayOfYear = gregorian.ToJd(scal.Date{date.Year + 78, 3, 22})
		daysInMonth1 = 30
	}

	// Add onto the jd of the first day of the year the number of days required
	// Calculate the number of days in the months before the required month
	// Then add on the required days
	// The first month has 30 or 31 days depending on if it is a Leap Year (determined above)
	// The second to sixth months have 31 days each
	// The seventh to twelfth months have 30 days each
	// Note: could be expressed more efficiently, but I think this is clearer
	var jd int
	if date.Month == 1 {
		jd = jdFirstDayOfYear + int(date.Day) - 1
	} else if date.Month <= 6 {
		jd = jdFirstDayOfYear + daysInMonth1 + (int(date.Month)-2)*31 + int(date.Day) - 1
	} else { // date.Month > 6
		jd = jdFirstDayOfYear + daysInMonth1 + 5*31 + (int(date.Month)-7)*30 + int(date.Day) - 1
	}
	return jd

}

func JdTo(jd int) scal.Date {
	var year, month, day int

	// The calendar is closely synchronized to the Gregorian Calendar, always starting on the same day
	// We can use this and the regular sequence of days in months to do a simple conversion by finding
	// what day in the Gregorian year the Julian Day number is, converting this to the day in the
	// Indian year and subtracting off the required number of months and days to get the final date

	gDate := gregorian.JdTo(jd)
	jdGregorianFirstDayOfYear := gregorian.ToJd(scal.Date{gDate.Year, 1, 1})
	gregorianDayOfYear := jd - jdGregorianFirstDayOfYear + 1

	// There is a fixed 78 year difference between year numbers, but the years do not exactly match up,
	// there is a fixed 80 day difference between the first day of the year, if the Gregorian day of
	// the year is 80 or less then the equivalent Indian day actually falls in the preceding    year
	if gregorianDayOfYear > 80 {
		year = gDate.Year - 78
	} else {
		year = gDate.Year - 79
	}

	var daysInMonth1 int
	// If it is a leap year then the first month has 31 days, otherwise 30.
	if IsLeap(year) {
		daysInMonth1 = 31
	} else {
		daysInMonth1 = 30
	}

	// The Indian year always starts 80 days after the Gregorian year, calculate the Indian day of
	// the year, taking into account if it falls into the previous Gregorian year
	var indianDayOfYear int
	if gregorianDayOfYear > 80 {
		indianDayOfYear = gregorianDayOfYear - 80
	} else {
		indianDayOfYear = gregorianDayOfYear + daysInMonth1 + 5*31 + 6*30 - 80
	}

	// Then simply remove the whole months from the day of the year and you are left with the day of month
	if indianDayOfYear <= daysInMonth1 {
		month = 1
		day = indianDayOfYear
	} else if indianDayOfYear <= daysInMonth1+5*31 {
		month = (indianDayOfYear-daysInMonth1-1)/31 + 2
		day = indianDayOfYear - daysInMonth1 - (month-2)*31
	} else {
		month = (indianDayOfYear-daysInMonth1-5*31-1)/30 + 7
		day = indianDayOfYear - daysInMonth1 - 5*31 - (month-7)*30
	}
	return scal.Date{year, uint8(month), uint8(day)}
}

func GetMonthLen(year int, month uint8) uint8 {
	if month == 1 {
		if IsLeap(year) {
			return 31
		} else {
			return 30
		}
	}
	if 2 <= month && month <= 6 {
		return 31
	}
	return 30
}
