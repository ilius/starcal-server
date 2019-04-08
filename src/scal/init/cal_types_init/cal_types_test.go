package cal_types_init

import "testing"

import "scal"
import "github.com/ilius/libgostarcal/cal_types"

func TestConvert(t *testing.T) {
	t.Log(cal_types.CalTypesMap["gregorian"])
	gdate := lib.Date{2016, 1, 1}
	jdate, err := cal_types.Convert(gdate, "gregorian", "jalali")
	if err != nil {
		t.Error(err)
	}
	if jdate.String() != "1394/10/11" {
		t.Error("Wrong: jdate =", jdate)
	}
	//t.Logf("%v => %v\n", gdate, jdate)
}
func TestToJd(t *testing.T) {
	gdate := lib.Date{2016, 1, 1}
	jd, err := cal_types.ToJd(gdate, "gregorian")
	if err != nil {
		t.Error(err)
	}
	//t.Log("jd =", jd)
	gdate2, err2 := cal_types.JdTo(jd, "gregorian")
	if err2 != nil {
		t.Error(err2)
	}
	//t.Log("gdate2 =", gdate2)
	if gdate2.String() != "2016/01/01" {
		t.Error("Wrong: gdate2 =", gdate2)
	}
}
