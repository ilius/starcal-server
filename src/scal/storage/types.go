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
	IsNotFound(error) bool
	Insert(model hasCollection) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	First(scal.M, string, hasCollection) error
	FindCount(string, scal.M) (int, error)
	FindAll(string, scal.M, interface{}) error
	PipeAll(string, []scal.M, interface{}) error
	PipeIter(string, []scal.M) <-chan scal.MErr
}
