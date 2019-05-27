package api_v1

import (
	"fmt"
	"scal"
	"strings"

	"scal/settings"

	"github.com/globalsign/mgo/bson"
	. "github.com/ilius/ripo"
)

func ObjectIdFromRequest(req Request, name string, sources ...FromX) (*string, error) {
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
	return objIdHex, nil
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
	log.Warn("No default page limit for ", handlerName)
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

func GetPageExStartId(req Request) (*string, error) {
	return ObjectIdFromRequest(req, "exStartId", FromBody, FromForm, FromEmpty)
}

func GetReverseOrder(req Request) (bool, error) {
	reverseOrder, err := req.GetBool("reverseOrder", FromBody, FromForm, FromEmpty)
	if err != nil {
		return false, err
	}
	return *reverseOrder, nil
}

func GetPageOptions(req Request) (*scal.PageOptions, error) {
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
	pageOpts := &scal.PageOptions{
		ReverseOrder: reverseOrder,
		Limit:        limit,
	}
	if exStartId != nil {
		pageOpts.ExStartId = exStartId
	}
	return pageOpts, nil
}
