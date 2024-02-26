package api_v1

import (
	"fmt"

	"github.com/ilius/starcal-server/pkg/scal"
	"github.com/ilius/starcal-server/pkg/scal/storage"

	"github.com/ilius/mgo/bson"
)

func NewPipelines(db storage.Database, collection string) *MongoPipelines {
	return &MongoPipelines{
		db:         db,
		collection: collection,
		match:      scal.M{},
	}
}

type AccessModelInterface interface {
	CanRead(email string) bool
}

type GroupByInterface interface {
	AddFromFirst(string, string) GroupByInterface
	AddFromLast(string, string) GroupByInterface
}

type MongoPipelines struct {
	db         storage.Database
	collection string

	match     scal.M
	limit     int
	pipelines []any
	trail     []scal.M
}

type MongoGroupBy struct {
	groupByKey string
	fields     scal.M
}

func (mg *MongoGroupBy) AddFromFirst(key string, alias string) GroupByInterface {
	mg.fields[alias] = scal.M{
		"$first": "$" + key,
	}
	return mg
}

func (mg *MongoGroupBy) AddFromLast(key string, alias string) GroupByInterface {
	mg.fields[alias] = scal.M{
		"$last": "$" + key,
	}
	return mg
}

func (mg *MongoGroupBy) prepare() scal.M {
	if len(mg.fields) == 0 {
		return nil
	}
	value := scal.M{"_id": "$" + mg.groupByKey}
	for fieldKey, fieldValue := range mg.fields {
		value[fieldKey] = fieldValue
	}
	return scal.M{"$group": value}
}

func (m *MongoPipelines) MatchValue(key string, value any) {
	m.match[key] = value
}

func (m *MongoPipelines) MatchGreaterThan(key string, value any) {
	m.match[key] = scal.M{
		"$gt": value,
	}
}

func (m *MongoPipelines) NewMatchGreaterThan(key string, value any) {
	m.pipelines = append(m.pipelines, scal.M{
		"$match": scal.M{
			"$gt": value,
		},
	})
}

func (m *MongoPipelines) Sort(key string, ascending bool) {
	value := -1
	if ascending {
		value = 1
	}
	m.pipelines = append(m.pipelines, scal.M{"$sort": scal.M{key: value}})
}

func (m *MongoPipelines) SortAtTheEnd(key string, ascending bool) {
	value := -1
	if ascending {
		value = 1
	}
	m.trail = append(m.trail, scal.M{"$sort": scal.M{key: value}})
}

func (m *MongoPipelines) Limit(limit int) {
	m.limit = limit
}

func (m *MongoPipelines) AppendLimit(limit int) {
	m.pipelines = append(m.pipelines, scal.M{"$limit": limit})
}

// returns "$gt" or "$lt"
func (m *MongoPipelines) startIdOperation(o *scal.PageOptions) string {
	if o.ReverseOrder {
		return "$lt"
	}
	return "$gt"
}

func (m *MongoPipelines) SetPageOptions(o *scal.PageOptions) {
	if o.ExStartId != nil {
		m.match["_id"] = scal.M{
			m.startIdOperation(o): bson.ObjectIdHex(*o.ExStartId),
		}
	}
	m.Sort("_id", !o.ReverseOrder)
	m.Limit(o.Limit)
	m.SortAtTheEnd("_id", !o.ReverseOrder)
}

func (m *MongoPipelines) AddEventLookupMetaAccess(
	email string,
	localField string,
) {
	m.pipelines = append(m.pipelines, []any{
		scal.M{"$lookup": scal.M{
			"from":         storage.C_eventMeta,
			"localField":   localField,
			"foreignField": "_id",
			"as":           "meta",
		}},
		scal.M{"$unwind": "$meta"},
		scal.M{"$match": scal.M{
			"$or": []scal.M{
				{"meta.ownerEmail": email},
				{"meta.isPublic": true},
				{"meta.accessEmails": email},
			},
		}},
	}...)
}

func (m *MongoPipelines) AddEventGroupAccess(email string) {
	m.match["groupId"] = scal.M{
		"$or": []scal.M{
			{"ownerEmail": email},
			{"isPublic": true},
			{"accessEmails": email},
		},
	}
}

// func (model *EventGroupModel) GetAccessCond(email string) scal.M {
// 		return
// }

func (m *MongoPipelines) GroupBy(key string) GroupByInterface {
	groupBy := &MongoGroupBy{
		groupByKey: key,
		fields:     scal.M{},
	}
	m.pipelines = append(m.pipelines, groupBy)
	return groupBy
}

func (m *MongoPipelines) Lookup(from string, localField string, foreignField string, as string) {
	m.pipelines = append(m.pipelines, scal.M{"$lookup": scal.M{
		"from":         from,
		"localField":   localField,
		"foreignField": foreignField,
		"as":           as,
	}})
}

func (m *MongoPipelines) Unwind(key string) {
	m.pipelines = append(m.pipelines, scal.M{"$unwind": "$" + key})
}

func (m *MongoPipelines) prepare() []scal.M {
	pipelines := []scal.M{}
	if len(m.match) > 0 {
		pipelines = append(pipelines, scal.M{
			"$match": m.match,
		})
	}
	if m.limit > 0 {
		pipelines = append(pipelines, scal.M{"$limit": m.limit})
	}
	for _, p := range m.pipelines {
		switch pt := p.(type) {
		case scal.M:
			if len(pt) == 0 {
				fmt.Println("MongoPipelines: map is empty")
				break
			}
			pipelines = append(pipelines, pt)
		case *MongoGroupBy:
			m := pt.prepare()
			if m == nil {
				fmt.Println("MongoPipelines: GroupBy is empty")
				break
			}
			pipelines = append(pipelines, m)
		default:
			panic(fmt.Errorf("invalid type %T", p))
		}
	}
	for _, p := range m.trail {
		pipelines = append(pipelines, p)
	}
	// for _, p := range pipelines {
	// 	b, _ := json.MarshalIndent(p, "", "    ")
	// 	fmt.Println(string(b))
	// }
	return pipelines
}

func (m *MongoPipelines) All(results any) error {
	pipelines := m.prepare()
	return storage.PipeAll(
		m.db,
		m.collection,
		&pipelines,
		results,
	)
}

func (m *MongoPipelines) Iter() (
	next func(result any) error,
	close func(),
) {
	pipelines := m.prepare()
	return m.db.PipeIter(
		m.collection,
		&pipelines,
	)
}
