package api_v1

import (
	"fmt"
	"log"
	"scal"
	"strings"

	"scal/settings"

	. "github.com/ilius/ripo"
	"github.com/globalsign/mgo/bson"
)

func ObjectIdFromRequest(req Request, name string, sources ...FromX) (*bson.ObjectId, error) {
	objIdHex, err := req.GetString(name, sources...)
	if err != nil {
		return nil, err
	}
	if *objIdHex == "" {
		return nil, nil
	}
	if !bson.IsObjectIdHex(*objIdHex) { // to avoid panic!
		return nil, NewError(InvalidArgument, fmt.Sprintf("invalid '%s'", name), nil)
	}
	objId := bson.ObjectIdHex(*objIdHex)
	return &objId, nil
}

func GetDefaultPageLimit(req Request) int {
	handlerNameFull := req.HandlerName()
	handlerNameParts := strings.Split(handlerNameFull, "/")
	handlerName := handlerNameParts[len(handlerNameParts)-1]
	// for example: handlerName == "api_v1.GetMyEventList"
	limit := settings.API_PAGE_LIMITS[handlerName]
	if limit > 0 {
		return limit
	}
	log.Println("No default page limit for", handlerName)
	return settings.API_PAGE_LIMIT_DEFAULT
}

func GetPageLimit(req Request) (int, error) {
	limitDefault := GetDefaultPageLimit(req)

	limit, err := req.GetIntDefault("limit", 0)
	if err != nil || limit <= 0 {
		return limitDefault, err
	}

	limitMax := int(float64(limitDefault) * settings.API_PAGE_LIMIT_MAX_RATIO)
	if limit > limitMax {
		limit = limitMax
	}
	return limit, nil
}

func GetPageExStartId(req Request) (*bson.ObjectId, error) {
	return ObjectIdFromRequest(req, "exStartId", FromBody, FromForm, FromEmpty)
}

func GetReverseOrder(req Request) (bool, error) {
	reverseOrder, err := req.GetBool("reverseOrder", FromBody, FromForm, FromEmpty)
	if err != nil {
		return false, err
	}
	return *reverseOrder, nil
}

type PageOptions struct {
	ExStartId    *bson.ObjectId
	ReverseOrder bool
	Limit        int
}

// returns "$gt" or "$lt"
func (o *PageOptions) StartIdOperation() string {
	if o.ReverseOrder {
		return "$lt"
	}
	return "$gt"
}

func (o *PageOptions) AddStartIdCond(cond scal.M) {
	if o.ExStartId != nil {
		cond["_id"] = scal.M{
			o.StartIdOperation(): *(o.ExStartId),
		}
	}
}

// returns "_id" or "-_id"
func (o *PageOptions) SortBy() string {
	if o.ReverseOrder {
		return "-_id"
	}
	return "_id"
}

// for pipelines
func (o *PageOptions) SortByMap() scal.M {
	orderInt := 1
	if o.ReverseOrder {
		orderInt = -1
	}
	return scal.M{"$sort": scal.M{"_id": orderInt}}
}

func GetPageOptions(req Request) (*PageOptions, error) {
	exStartId, err := GetPageExStartId(req)
	if err != nil {
		return nil, err
	}
	reverseOrder, err := GetReverseOrder(req)
	if err != nil {
		return nil, err
	}
	limit, err := GetPageLimit(req)
	if err != nil {
		return nil, err
	}
	return &PageOptions{
		ExStartId:    exStartId,
		ReverseOrder: reverseOrder,
		Limit:        limit,
	}, nil
}
