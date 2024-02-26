package storage

import (
	"github.com/ilius/starcal-server/pkg/scal"

	"github.com/ilius/mgo/bson"
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
	Equals(key string, value any) Condition
	IdEquals(key string, value string) Condition
	Includes(key string, value any) Condition
	GreaterThan(key string, value any) Condition
	LessThan(key string, value any) Condition
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
	IsNotFound(err error) bool
	Insert(model hasCollection) error
	InsertMany(collection string, models []any) error
	Update(model hasCollectionUniqueM) error
	Upsert(model hasCollectionUniqueM) error
	Remove(model hasCollectionUniqueM) error
	RemoveAll(collection string, cond scal.M) (count int, err error)
	Get(model hasCollectionUniqueM) error
	First(cond scal.M, sortBy string, model hasCollection) error
	FindCount(collection string, cond scal.M) (count int, err error)
	FindAll(result any, input *FindInput) error
	PipeIter(collection string, pipeline *[]scal.M) (
		next func(result any) error,
		close func(),
	)
}
