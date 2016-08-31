package utils

import "time"

type Event interface {
    Id() int
    Location() *time.Location
}

type NilEvent struct {}
func (self NilEvent) Location() *time.Location {
    return time.Now().Location()
}
func (self NilEvent) Id() int {
    return -1
}
func (self NilEvent) String() string {
    return "NilEvent{}"
}


