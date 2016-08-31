package scal

import "fmt"

type Date struct {
    Year int
    Month int
    Day int
}
func (self Date) String() string {
    return fmt.Sprintf("%.4d/%.2d/%.2d", self.Year, self.Month, self.Day)
}
func (self Date) Repr() string {
    return fmt.Sprintf("scal.Date{%d, %d, %d}", self.Year, self.Month, self.Day)
}

