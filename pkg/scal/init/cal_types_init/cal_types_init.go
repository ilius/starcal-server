// Package cal_types_init loads / registers calendar types

package cal_types_init

import (
	_ "github.com/ilius/libgostarcal/cal_types/ethiopian"
	"github.com/ilius/libgostarcal/cal_types/hijri"
	"github.com/ilius/libgostarcal/cal_types/jalali"
	"github.com/ilius/starcal-server/pkg/scal/settings"
)

func init() {
	jalali.SetAlgorithm2820(settings.JALALI_ALGORITHM_2820)
	hijri.SetUseMonthData(settings.HIJRI_USE_MONTH_DATA)
}
