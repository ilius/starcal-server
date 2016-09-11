package storage

import "time"

//import "mgo"
import "gopkg.in/mgo.v2-unstable"
//import "gopkg.in/mgo.v2-unstable/bson"

const (
    MongoHost = "127.0.0.1:27017"
    MongoDbName = "starcal"
    MongoUsername = ""
    MongoPassword = ""
)


func GetDB() (*mgo.Database, error) {
    mongoDBDialInfo := &mgo.DialInfo{
        Addrs:    []string{MongoHost},
        Timeout:  2 * time.Second,
        Database: MongoDbName,
        Username: MongoUsername,
        Password: MongoPassword,
    }


    // Create a session which maintains a pool of socket connections
    // to our MongoDB.
    mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
    if err != nil {
        return nil, err
    }

    // Reads may not be entirely up-to-date, but they will always see the
    // history of changes moving forward, the data read will be consistent
    // across sequential queries in the same session, and modifications made
    // within the session will be observed in following queries (read-your-writes).
    // http://godoc.org/labix.org/v2/mgo#Session.SetMode
    mongoSession.SetMode(mgo.Monotonic, true)

    return mongoSession.DB(MongoDbName), nil
}








