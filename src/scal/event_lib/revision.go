package event_lib

import "time"

type EventRevisionModel struct {
    UserId int
    EventId string
    EventType string
    Hash string
    Time time.Time
    //InvitedUserIds []int
}





