package storage

import (
	"fmt"
	"time"

	mgo "github.com/ilius/mgo"
	"github.com/ilius/starcal-server/pkg/scal/settings"
)

var db *MongoDatabase

func GetDB() (Database, error) {
	if db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}
	return db, nil
}

func InitDB() {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{settings.MONGO_HOST},
		Timeout:  2 * time.Second,
		Database: settings.MONGO_DB_NAME,
		Username: settings.MONGO_USERNAME,
		Password: settings.MONGO_PASSWORD,
	}

	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	mongoSession, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		panic(err)
	}

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	mongoSession.SetMode(mgo.Monotonic, true)

	db = &MongoDatabase{
		*mongoSession.DB(settings.MONGO_DB_NAME),
	}
}
