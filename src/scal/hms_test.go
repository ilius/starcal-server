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
	str := "12:01:01"
	obj, err := ParseHMS(str)
	//t.Log(obj, err)
	if err != nil {
		t.Error(err)
	}
	if obj.String() != str {
		t.Error("Failed to parse HMS:", obj)
	}
}

func TestParseDHMS(t *testing.T) {
	str := "90 12:01:01"
	obj, err := ParseDHMS(str)
	//t.Log(obj, err)
	if err != nil {
		t.Error(err)
	}
	if obj.String() != str {
		t.Errorf("Failed to parse DHMS: obj = %v", obj)
	}
}
