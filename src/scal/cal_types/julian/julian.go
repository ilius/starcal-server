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

package julian

import (
	"scal"
	"scal/cal_types"
)

// ###### Common Globals #######

const (
	Name        = "julian"
	Desc        = "Julian"
	Epoch       = 1721058
	MinMonthLen = 28
	MaxMonthLen = 32
	AvgYearLen  = 365.25
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

var monthLen = []int{
	31,
	28, // (leap years: 29)
	31,
	30, 31, 30,
	31, 31, 30,
	31, 30, 31,
}

var monthLenSum = []int{
	0, 31, 59,
	90, 120, 151,
	181, 212, 243,
	273, 304, 334,
	365,
}

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
	return year%4 == 0
}

func getYearDays(month int, leap bool) int {
	// month: 1..13
	ydays := monthLenSum[month-1]
	if leap && month < 3 {
		ydays -= 1
	}
	return ydays
}

func getMonthDayFromYdays(yDays int, leap bool) (int, int) {
	// yDays: int, number of days in year
	month := 1
	for month < 12 && yDays > getYearDays(month+1, leap) {
		month += 1
	}
	day := yDays - getYearDays(month, leap)
	return month, day
}

func ToJd(date scal.Date) int {
	quadCount, yMode := divmod(date.Year, 4)
	return (Epoch +
		1461*quadCount +
		365*yMode +
		getYearDays(date.Month, yMode == 0) +
		date.Day)
}

func JdTo(jd int) scal.Date {
	quadCount, quadDays := divmod(jd-Epoch, 1461)

	if quadDays == 0 {
		// first day of quad (and year)
		return scal.Date{4 * quadCount, 1, 1}
	}

	yMode, yDays := divmod(quadDays-1, 365)
	yDays += 1
	year := 4*quadCount + yMode
	month, day := getMonthDayFromYdays(yDays, yMode == 0)

	return scal.Date{year, month, day}
}

func GetMonthLen(year int, month int) int {
	if month == 2 {
		if IsLeap(year) {
			return 29
		}
		return 28
	}
	return monthLen[month-1]
}
