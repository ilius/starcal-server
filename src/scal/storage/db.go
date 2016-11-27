package storage

import (
	//"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"scal/settings"
	"time"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() bson.M
}

type MongoDatabase struct {
	mgo.Database
}

func (db MongoDatabase) Insert(model hasCollection) error {
	return db.C(model.Collection()).Insert(model)
}
func (db MongoDatabase) Update(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Update(
		model.UniqueM(),
		model,
	)
}
func (db MongoDatabase) Upsert(model hasCollectionUniqueM) error {
	_, err := db.C(model.Collection()).Upsert(
		model.UniqueM(),
		model,
	)
	return err
}
func (db MongoDatabase) Remove(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Remove(
		model.UniqueM(),
	)
}

//func (db MongoDatabase) Find(interface{})
func (db MongoDatabase) Get(model hasCollectionUniqueM) error {
	return db.C(model.Collection()).Find(
		model.UniqueM(),
	).One(model)
}
func (db MongoDatabase) First(
	cond bson.M,
	sortBy string,
	model hasCollection,
) error {
	return db.C(model.Collection()).Find(cond).Sort(sortBy).One(model)
}

func (db MongoDatabase) FindAll(
	colName string,
	cond bson.M,
	result interface{},
) error {
	return db.C(colName).Find(cond).All(result)
}
func (db MongoDatabase) PipeAll(
	colName string,
	pipeline []bson.M,
	result interface{},
) error {
	return db.C(colName).Pipe(pipeline).All(result)
}

func GetDB() (*MongoDatabase, error) {
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
		return nil, err
	}

	// Reads may not be entirely up-to-date, but they will always see the
	// history of changes moving forward, the data read will be consistent
	// across sequential queries in the same session, and modifications made
	// within the session will be observed in following queries (read-your-writes).
	// http://godoc.org/labix.org/v2/mgo#Session.SetMode
	mongoSession.SetMode(mgo.Monotonic, true)

	return &MongoDatabase{
		*mongoSession.DB(settings.MONGO_DB_NAME),
	}, nil
}
