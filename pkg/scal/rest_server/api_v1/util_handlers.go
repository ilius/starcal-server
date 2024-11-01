package api_v1

import (
	"net/http"

	"github.com/ilius/ripo"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "util",
		Map: RouteMap{
			"GetApiVersion": {
				Method:  "GET",
				Pattern: "api-version",
				Handler: GetApiVersion,
			},
		},
	})
}

func GetApiVersion(_ ripo.Request) (*ripo.Response, error) {
	header := http.Header{}
	header.Set("Content-Type", "text/plain; charset=UTF-8")
	return &ripo.Response{
		Header: header,
		Data:   "1",
	}, nil
}
