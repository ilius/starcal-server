package storage

import (
	"fmt"
	"scal"

	"github.com/globalsign/mgo/bson"
)

type MongoCondition struct {
	op    ConditionOperator
	parts []interface{}
}

func (c *MongoCondition) Equals(key string, value interface{}) Condition {
	c.parts = append(c.parts, bson.DocElem{Name: key, Value: value})
	return c
}

func (c *MongoCondition) IdEquals(key string, valueHex string) Condition {
	c.parts = append(c.parts, bson.DocElem{Name: key, Value: bson.ObjectIdHex(valueHex)})
	return c
}

func (c *MongoCondition) Includes(key string, value interface{}) Condition {
	// for Mongo, it's the same as Equals()
	c.parts = append(c.parts, bson.DocElem{Name: key, Value: value})
	return c
}

func (c *MongoCondition) GreaterThan(key string, value interface{}) Condition {
	c.parts = append(c.parts, bson.DocElem{
		Name: key,
		Value: scal.M{
			"$gt": value,
		},
	})
	return c
}

func (c *MongoCondition) LessThan(key string, value interface{}) Condition {
	c.parts = append(c.parts, bson.DocElem{
		Name: key,
		Value: scal.M{
			"$lt": value,
		},
	})
	return c
}

func (c *MongoCondition) SetPageOptions(o *scal.PageOptions) Condition {
	if c.op != AND {
		return c.NewSubCondition(AND).SetPageOptions(o)
	}
	if o.ExStartId != nil {
		if o.ReverseOrder {
			c.LessThan("_id", o.ExStartId)
		}
		c.GreaterThan("_id", o.ExStartId)
	}
	return c
}

func (c *MongoCondition) NewSubCondition(operator ConditionOperator) Condition {
	c2 := &MongoCondition{
		op: operator,
	}
	c.parts = append(c.parts, c2)
	return c2
}

func (c *MongoCondition) Prepare() bson.D {
	parts := bson.D{}
	for _, p := range c.parts {
		switch pt := p.(type) {
		case bson.DocElem:
			parts = append(parts, pt)
		case *MongoCondition:
			for _, elem := range pt.Prepare() {
				parts = append(parts, elem)
			}
		default:
			panic(fmt.Errorf("invalid type %T", p))
		}
	}
	if c.op == AND {
		return parts
	}
	if c.op == OR {
		maps := []bson.M{}
		for _, p := range parts {
			maps = append(maps, bson.M{
				p.Name: p.Value,
			})
		}
		return bson.D{
			{Name: "$or", Value: maps},
		}
	}
	panic(fmt.Errorf("invalid c.op = %v", c.op))
	return nil
}
