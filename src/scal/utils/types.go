package utils

import "time"

type Event interface {
    GetLoc() *time.Location
}


