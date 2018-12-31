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

func (NilEvent) String() string {
	return "NilEvent{}"
}
func (NilEvent) Location() *time.Location {
	return time.Now().Location()
}
func (NilEvent) CalType() cal_types.CalType {
	calType, err := cal_types.GetCalType("gregorian")
	if err != nil {
		//log.Error(log)
	}
	return calType
}

func (NilEvent) Id() string {
	return "Nil"
}
func (NilEvent) Summary() string {
	return "Nil"
}
func (NilEvent) Description() string {
	return ""
}
func (NilEvent) Icon() string {
	return ""
}
func (NilEvent) NotifyBefore() int {
	return 0
}
