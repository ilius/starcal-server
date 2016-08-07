package main

import "fmt"

import . "scal"
import _ "scal/cal_types/loader"
import . "scal/cal_types"

//import "scal/cal_types/jalali"
//import "scal/cal_types/gregorian"
//import "scal/cal_types/indian_national"

func main() {
    //jalali.TestIsLeap(2000, 2020)
    //jalali.TestToJd(2000, 2010)
    //jalali.TestConvert(2000, 2010)
    //gregorian.TestIsLeap(2000, 2020)
    //gregorian.TestToJd(2000, 2020)
    //gregorian.TestConvert(2000, 2020)
    //indian_national.TestIsLeap(1930, 1950)
    //indian_national.TestToJd(1930, 1950)
    //indian_national.TestConvert(1930, 1950)
    
    //fmt.Println(CalTypesList)
    fmt.Println(CalTypesMap)
    //fmt.Println(CalTypesMap["gregorian"])
    //fmt.Println(CalTypesMap["jalali"])
    gdate := Date{2016, 1, 1}
    jdate, err := Convert(gdate, "gregorian", "indian_national")
    fmt.Printf("%v => %v   error=%v\n", gdate, jdate, err)
}

