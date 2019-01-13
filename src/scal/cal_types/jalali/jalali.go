// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
// Copyright (C) 2007 Mehdi Bayazee <Bayazee@Gmail.com>
// Copyright (C) 2001 Roozbeh Pournader <roozbeh@sharif.edu>
// Copyright (C) 2001 Mohammad Toossi <mohammad@bamdad.org>
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

// Iranian (Jalali) calendar:
// http://en.wikipedia.org/wiki/Iranian_calendar

package jalali

import "scal"
import . "scal/utils"
import "scal/cal_types"

// ###### Common Globals #######

const (
	Name  = "jalali"
	Desc  = "Jalali"
	Epoch = 1948321

	MinMonthLen uint8 = 29
	MaxMonthLen uint8 = 31

	AvgYearLen = 365.2425 // FIXME
)

var MonthNames = []string{
	"Farvardin", "Ordibehesht", "Khordad", "Teer", "Mordad", "Shahrivar",
	"Mehr", "Aban", "Azar", "Dey", "Bahman", "Esfand",
}
var MonthNamesAb = []string{
	"Far", "Ord", "Khr", "Tir", "Mor", "Shr",
	"Meh", "Abn", "Azr", "Dey", "Bah", "Esf",
}

// ###### Other Globals  #######

var monthLen = []uint8{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 30}
var monthLenSum = []int{0, 31, 62, 93, 124, 155, 186, 216, 246, 276, 306, 336, 366}

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

// Normal: esfand = 29 days
// Leap: esfand = 30 days
func IsLeap(year int) bool {
	// return true if year is leap, false otherwise
	// using 2820-years algorithm
	if year > 0 {
		year--
	}
	return (((year-473)%2820)*682)%2816 < 682
}

func ToJd(date scal.Date) int {
	// calculate Julian day from Jalali date
	// using 2820-years algorithm
	var epbase int
	if date.Year >= 0 {
		epbase = date.Year - 474
	} else {
		epbase = 473
	}
	epyear := 474 + epbase%2820
	mm := int(date.Month - 1)
	jd := int(date.Day) +
		mm*30 + IntMin(6, mm) +
		(epyear*682-110)/2816 +
		(epyear-1)*365 +
		epbase/2820*1029983 +
		Epoch - 1
	return jd
}

func JdTo(jd int) scal.Date {
	// calculate Jalali date from Julian day
	// using 2820-years algorithm
	deltaDays := jd - ToJd(scal.Date{475, 1, 1})
	cycle := deltaDays / 1029983
	cyear := deltaDays % 1029983
	var ycycle int
	if cyear == 1029982 {
		ycycle = 2820
	} else {
		ycycle = (2134*(cyear/366)+2816*(cyear%366)+2815)/1028522 + cyear/366 + 1
	}
	year := 2820*cycle + ycycle + 474
	if year <= 0 {
		year--
	}
	yday := jd - ToJd(scal.Date{year, 1, 1}) + 1
	month := uint8(BisectLeft(monthLenSum, yday))
	day := uint8(yday - monthLenSum[month-1])
	return scal.Date{year, month, day}
}

func GetMonthLen(year int, month uint8) uint8 {
	if month == 12 {
		if IsLeap(year) {
			return 30
		} else {
			return 29
		}
	} else {
		return monthLen[month-1]
	}
}
