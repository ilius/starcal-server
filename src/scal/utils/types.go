package utils

import "time"
import "scal/cal_types"

type Event interface {
	String() string
	Location() *time.Location
	CalType() cal_types.CalType

	Id() string
	Summary() string
	Description() string
	Icon() string
	NotifyBefore() int
}

type NilEvent struct{}

func (self NilEvent) String() string {
	return "NilEvent{}"
}
func (self NilEvent) Location() *time.Location {
	return time.Now().Location()
}
func (self NilEvent) CalType() cal_types.CalType {
	calType, err := cal_types.GetCalType("gregorian")
	if err != nil {
		//log.Error(log)
	}
	return calType
}

func (self NilEvent) Id() string {
	return "Nil"
}
func (self NilEvent) Summary() string {
	return "Nil"
}
func (self NilEvent) Description() string {
	return ""
}
func (self NilEvent) Icon() string {
	return ""
}
func (self NilEvent) NotifyBefore() int {
	return 0
}
