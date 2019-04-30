package storage

import (
	"scal"

	"github.com/globalsign/mgo/bson"
)

type hasCollection interface {
	Collection() string
}

type hasCollectionUniqueM interface {
	Collection() string
	UniqueM() scal.M
}

type ConditionOperator string

const (
	AND ConditionOperator = "and"
	OR  ConditionOperator = "or"
)

type Condition interface {
	Equals(key string, value interface{}) Condition
	Includes(key string, value interface{}) Condition
	GreaterThan(key string, value interface{}) Condition
	LessThan(key string, value interface{}) Condition
	SetPageOptions(o *scal.PageOptions) Condition
	NewSubCondition(operator ConditionOperator) Condition
	Prepare() bson.D
}

type FindInput struct {
	Collection   string
	Condition    Condition
	SortBy       string
	ReverseOrder bool
	Limit        int
	Fields       []string
}

type Database interface {
	NewCondition(operator ConditionOperator) Condition
	IsNotFound(error) bool
	Insert(model hasCollection) error
	InsertMany(collection string, models []interface{}) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	Get(model hasCollectionUniqueM) error
	First(scal.M, string, hasCollection) error
	FindCount(string, scal.M) (int, error)
	FindAll(result interface{}, input *FindInput) error
	PipeIter(collection string, pipeline *[]scal.M) (
		next func(result interface{}) error,
		close func(),
	)
}
