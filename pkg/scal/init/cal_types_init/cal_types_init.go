// Package cal_types_init loads / registers calendar types

package cal_types_init

import (
	_ "github.com/ilius/libgostarcal/cal_types/ethiopian"
	_ "github.com/ilius/libgostarcal/cal_types/gregorian"
	_ "github.com/ilius/libgostarcal/cal_types/gregorian_proleptic"
	_ "github.com/ilius/libgostarcal/cal_types/hijri"
	_ "github.com/ilius/libgostarcal/cal_types/indian_national"
	"github.com/ilius/libgostarcal/cal_types/jalali"
	_ "github.com/ilius/libgostarcal/cal_types/julian"
	"github.com/ilius/starcal-server/pkg/scal/settings"
)

func init() {
	jalali.SetAlgorithm2820(settings.JALALI_ALGORITHM_2820)
}
