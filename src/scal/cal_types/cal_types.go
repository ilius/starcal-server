// -*- coding: utf-8 -*-
//
// Copyright (C) Saeed Rasooli <saeed.gnu@gmail.com>
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program. If not, see <http://www.gnu.org/licenses/gpl.txt>.
// Also avalable in /usr/share/common-licenses/GPL on Debian systems
// or /usr/share/licenses/common/GPL3/license.txt on ArchLinux


package cal_types

import "errors"

import "scal"

type CalType struct {
    Name string
    Desc string
    Epoch int
    MinMonthLen int
    MaxMonthLen int
    AvgYearLen float64
    MonthNames []string
    MonthNamesAb []string
    IsLeap func(year int) bool
    ToJd func(date scal.Date) int
    JdTo func(jd int) scal.Date
    GetMonthLen func(year int, month int) int
}

var CalTypesList []CalType
var CalTypesMap = make(map[string]CalType)

func RegisterCalType(
    Name string,
    Desc string,
    Epoch int,
    MinMonthLen int,
    MaxMonthLen int,
    AvgYearLen float64,
    MonthNames []string,
    MonthNamesAb []string,
    IsLeap func(year int) bool,
    ToJd func(date scal.Date) int,
    JdTo func(jd int) scal.Date,
    GetMonthLen func(year int, month int) int,
) {
    calType := CalType{
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
    }
    CalTypesList = append(CalTypesList, calType)
    CalTypesMap[Name] = calType
        
}

func Convert(date scal.Date, fromTypeName string, toTypeName string) (scal.Date, error) {
    fromType, fromOk := CalTypesMap[fromTypeName]
    toType, toOk := CalTypesMap[toTypeName]
    if !fromOk {
        return scal.Date{},
            errors.New("invalid calendar type '" + fromTypeName + "'")
    }
    if !toOk {
        return scal.Date{},
            errors.New("invalid calendar type '" + toTypeName + "'")
    }
    return toType.JdTo(fromType.ToJd(date)), nil
}














