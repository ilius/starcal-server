package event_lib

import (
    "testing"
    "time"

    _ "scal/cal_types/gregorian"
    _ "scal/cal_types/jalali"

    "scal/storage"

    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

const (
    MongoHost = "127.0.0.1:27017"
    MongoDbName = "starcal"
    MongoUsername = ""
    MongoPassword = ""
)


func TestMongoInsertEvents(t *testing.T) {
    mongoDBDialInfo := &mgo.DialInfo{
        Addrs:    []string{MongoHost},
        Timeout:  60 * time.Second,
        Database: MongoDbName,
        Username: MongoUsername,
        Password: MongoPassword,
    }


    // Create a session which maintains a pool of socket connections
    // to our MongoDB.
    mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
    if err != nil {
        t.Error(err)
    }

    // Reads may not be entirely up-to-date, but they will always see the
    // history of changes moving forward, the data read will be consistent
    // across sequential queries in the same session, and modifications made
    // within the session will be observed in following queries (read-your-writes).
    // http://godoc.org/labix.org/v2/mgo#Session.SetMode
    mongoSession.SetMode(mgo.Monotonic, true)

    db := mongoSession.DB(MongoDbName)
	col_events := db.C(storage.C_eventData)

    //err := col_events.Find(nil).All(&events)
    
    now := time.Now()
    //startTime := now.Add(time.Duration(-3600) * time.Second)
    //startTime := now.Add(-3600 * time.Second)
    startTime := now.Add(-2 * time.Hour)
    
    endTime := now
    
    eventModel := TaskEventModel {
        BaseEventModel: BaseEventModel {
            Id: bson.NewObjectId(),
            Summary: "test task",
            CalType: "jalali",
            TimeZone: "UTC",
            TimeZoneEnable: false,
            NotifyBefore: 0,
        },
        StartTime: startTime,
        EndTime: endTime,
    }
    err = col_events.Insert(eventModel)
    if err != nil {
        t.Log(err)
    }
    event, err2 := eventModel.GetEvent()
    if err2 != nil {
        t.Error(err2)
    }
    t.Log(event)
    t.Log("Event Type:", event.Type())
    t.Log("Event Location:", event.Location())
    t.Log("loc==nil:", event.loc == nil)
    /*
    eventModel2 := AllDayTaskEventModel {
        BaseEventModel: BaseEventModel {
            Type: "allDayTask",
            Summary: "test all-day task",
            CalType: "jalali",
            //TimeZoneEnable: false,
            NotifyBefore: 0,
        },
        StartJd: todayJd - 3,
        EndJd: todayJd,
    }
    err = col_events.Insert(eventModel2)
    if err != nil {
        t.Log(err)
    }
    */


}








