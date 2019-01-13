// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
// Used code from http://code.google.com/p/ethiocalendar/
//                Copyright (C) 2008-2009 Yuji DOI <yuji5296@gmail.com>
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

package ethiopian

import (
	"scal"
	"scal/cal_types"
	"scal/utils"
)

// ###### Common Globals #######

const (
	Name  = "ethiopian"
	Desc  = "Ethiopian"
	Epoch = 1724235

	MinMonthLen uint8 = 30
	MaxMonthLen uint8 = 36

	AvgYearLen = 365.25
)

var MonthNames = []string{
	"Meskerem", "Tekimt", "Hidar",
	"Tahsas", "Ter", "Yekoutit",
	"Meyabit", "Meyaziya", "Genbot",
	"Sene", "Hamle", "Nahse",
}
var MonthNamesAb = []string{
	"Meskerem", "Tekimt", "Hidar",
	"Tahsas", "Ter", "Yekoutit",
	"Meyabit", "Meyaziya", "Genbot",
	"Sene", "Hamle", "Nahse",
} // FIXME

// ###### Other Globals  #######

var monthLens = []uint8{
	30, 30, 30,
	30, 30, 30,
	30, 30, 30,
	30, 30, 35,
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
	return (year+1)%4 == 0
}

func ToJd(date scal.Date) int {
	return Epoch +
		365*(date.Year-1) + date.Year/4 +
		(int(date.Month-1))*30 +
		int(date.Day) - 15
}

func JdTo(jd int) scal.Date {
	quad := (jd - Epoch) / 1461
	dquad := (jd - Epoch) % 1461
	yindex := utils.IntMin(3, dquad/365)
	year := quad*4 + yindex + 1

	yearday := jd - ToJd(scal.Date{year, 1, 1})
	month := yearday/30 + 1
	day := yearday%30 + 1

	if month == 13 {
		month -= 1
		day += 30
	}
	if month == 12 {
		mLen := 35
		if IsLeap(year) {
			mLen += 1
		}
		if day > mLen {
			year += 1
			month = 1
			day -= mLen
		}
	}
	return scal.Date{year, uint8(month), uint8(day)}
}

func GetMonthLen(year int, month uint8) uint8 {
	if month == 12 {
		if IsLeap(year) {
			return 36
		} else {
			return 35
		}
	} else {
		return monthLens[month-1]
	}
}
