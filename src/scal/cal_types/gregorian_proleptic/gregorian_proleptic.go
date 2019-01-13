// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
// Using libkal code
//        The 'libkal' library for date conversion:
//        Copyright (C) 1996-1998 Petr Tomasek <tomasek@etf.cuni.cz>
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

package gregorian_proleptic

import (
	"scal"
	"scal/cal_types"
)

// ###### Common Globals #######

const (
	Name        = "julian"
	Desc        = "Julian"
	Epoch       = 1721426
	MinMonthLen = 29
	MaxMonthLen = 31
	AvgYearLen  = 365.2425 // FIXME
)

var MonthNames = []string{
	"January", "February", "March",
	"April", "May", "June",
	"July", "August", "September",
	"October", "November", "December",
}
var MonthNamesAb = []string{
	"Jan", "Feb", "Mar",
	"Apr", "May", "Jun",
	"Jul", "Aug", "Sep",
	"Oct", "Nov", "Dec",
}

// #############################

var monthLen = []uint8{
	31,
	28, // (leap years: 29)
	31,
	30, 31, 30,
	31, 31, 30,
	31, 30, 31,
}

/*var monthLenSum = []uint8{
	0, 31, 59,
	90, 120, 151,
	181, 212, 243,
	273, 304, 334,
	365,
}*/

func divmod(a int, b int) (int, int) {
	return a / b, a % b
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
	if year < 1 {
		year += 1
	}
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func ToJd(date scal.Date) int {
	/*
	   Formula from The Calendar FAQ by Claus Tondering
	   http://www.tondering.dk/claus/cal/node3.html#SECTION003161000000000000000
	   NOTE: Coded from scratch from mathematical formulas, not copied from
	   the Boost licensed source code

	   If year is -ve then is BC. In Gregorian there is no year 0,
	   but the maths is easier if we pretend there is,
	   so internally year of -1 = 1BC = 0 internally
	*/

	a := 0
	if date.Month < 3 {
		a = 1
	}

	y := date.Year + 4800 - a
	if date.Year < 1 {
		y += 1
	}

	m := int(date.Month) + 12*a - 3

	return (365*y + y/4 - y/100 + y/400 - 32045 +
		(153*m+2)/5 + int(date.Day))
}

func JdTo(jd int) scal.Date {
	/*
	   Formula from The Calendar FAQ by Claus Tondering
	   http://www.tondering.dk/claus/cal/node3.html#SECTION003161000000000000000
	   NOTE: Coded from scratch from mathematical formulas, not copied from
	   the Boost licensed source code
	*/
	a := jd + 32044
	b := (4*a + 3) / 146097
	c := a - 146097*b/4
	d := (4*c + 3) / 1461
	e := c - 1461*d/4
	m := (5*e + 2) / 153
	day := uint8(e - (153*m+2)/5 + 1)
	month := uint8(m + 3 - 12*(m/10))
	year := 100*b + d - 4800 + (m / 10)
	// If year is -ve then is BC. In Gregorian there is no year 0,
	// but the maths is easier if we pretend there is,
	// so internally year of 0 = 1BC = -1 outside
	if year < 1 {
		year -= 1
	}
	return scal.Date{year, month, day}
}

func GetMonthLen(year int, month uint8) uint8 {
	if month == 2 {
		if IsLeap(year) {
			return 29
		}
		return 28
	}
	return monthLen[month-1]
}
