package scal

import "testing"

func TestHMS_FloatHour(t *testing.T) {
	hms := HMS{12, 59, 5}
	fh := hms.GetFloatHour()
	hms2 := FloatHourToHMS(fh)
	if fh != 12.98472222222222 {
		t.Log("Wrong float hour: fh =", fh)
	}
	if hms2.String() != "12:59:05" {
		t.Log("Wrong HMS: hms2 =", hms2)
	}
	//t.Log("hms =", hms)
	//t.Log("fh =", fh)
	//t.Log("hms2 =", hms2)
}

func TestParseHMS(t *testing.T) {
	s := "12:01:01"
	hms, err := ParseHMS(s)
	//t.Log(hms, err)
	if err != nil {
		t.Error(err)
	}
	if hms.String() != "12:01:01" {
		t.Error("Wrong HMS: hms =", hms)
	}
}
