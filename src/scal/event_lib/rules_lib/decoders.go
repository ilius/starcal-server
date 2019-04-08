package rules_lib

import (
	lib "github.com/ilius/libgostarcal"
	"github.com/ilius/libgostarcal/interval"
	"github.com/ilius/libgostarcal/utils"
	//"fmt"
	"strconv"
)

var valueDecoders = map[string]func(string) (interface{}, error){
	T_string: func(value string) (interface{}, error) {
		return value, nil
	},
	T_int: func(value string) (interface{}, error) {
		v, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return 0, err
		}
		return int(v), nil
	},
	T_int_list: func(value string) (interface{}, error) {
		v, err := utils.ParseIntList(value)
		return v, err
	},
	T_int_range_list: func(value string) (interface{}, error) {
		intervalList, err := interval.ParseClosedIntervalList(value)
		if err != nil {
			return []int{}, err
		}
		intervalList, err = intervalList.Normalize()
		if err != nil {
			return []int{}, err
		}
		return utils.Int64ListToIntList(intervalList.Extract()), nil
	},
	T_float: func(value string) (interface{}, error) {
		v, err := strconv.ParseFloat(value, 64)
		return v, err
	},
	T_HMS: func(value string) (interface{}, error) {
		v, err := lib.ParseHMS(value)
		return v, err
	},
	T_DHMS: func(value string) (interface{}, error) {
		v, err := lib.ParseDHMS(value)
		return v, err
	},
	T_HMSRange: func(value string) (interface{}, error) {
		v, err := lib.ParseHMSRange(value)
		return v, err
	},
	T_Date: func(value string) (interface{}, error) {
		v, err := lib.ParseDate(value)
		return v, err
	},
	T_Date_list: func(value string) (interface{}, error) {
		v, err := lib.ParseDateList(value)
		return v, err
	},
	T_DateHMS: func(value string) (interface{}, error) {
		v, err := lib.ParseDateHMS(value)
		return v, err
	},
	T_Duration: func(value string) (interface{}, error) {
		v, err := utils.ParseDuration(value)
		return v, err
	},
}
