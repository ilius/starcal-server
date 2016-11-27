package storage

import (
	"gopkg.in/mgo.v2/bson"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() bson.M
}

type Database interface {
	Insert(model hasCollection) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	FindCount(string, bson.M) (int, error)
	FindAll(string, bson.M, interface{}) error
	First(bson.M, string, hasCollection) error
}
