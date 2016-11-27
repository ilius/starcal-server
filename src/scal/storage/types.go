package storage

import (
	"scal"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() scal.M
}

type Database interface {
	Insert(model hasCollection) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	FindCount(string, scal.M) (int, error)
	FindAll(string, scal.M, interface{}) error
	First(scal.M, string, hasCollection) error
}
