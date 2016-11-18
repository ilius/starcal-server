// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
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

package gregorian

import (
	"scal"
	"scal/cal_types"
	"time"
)

// ###### Common Globals #######

var Name = "gregorian"
var Desc = "Gregorian"

var Epoch = 1721426
var MinMonthLen = 29
var MaxMonthLen = 31
var AvgYearLen = 365.2425 // FIXME

var MonthNames = []string{
	"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December",
}
var MonthNamesAb = []string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// ###### Other Globals  #######

//var J0001 = Epoch
var J1970 = 2440588

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
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func ToJd(date scal.Date) int {
	t := time.Date(
		date.Year,
		time.Month(date.Month),
		date.Day,
		0, 0, 0,
		0,
		time.UTC,
	)
	return J1970 + int(t.Unix()/86400)
}

func JdTo(jd int) scal.Date {
	t := time.Unix(
		int64(86400*(jd-J1970)),
		0,
	)
	return scal.Date{
		t.Year(),
		int(t.Month()),
		t.Day(),
	}
}

func GetMonthLen(year int, month int) int {
	if month == 12 {
		return ToJd(scal.Date{year + 1, 1, 1}) - ToJd(scal.Date{year, 12, 1})
	} else {
		return ToJd(scal.Date{year, month + 1, 1}) - ToJd(scal.Date{year, month, 1})
	}
}
