package api_v1

import (
	. "github.com/ilius/ripo"
)

func init() {
	routeGroups = append(routeGroups, RouteGroup{
		Base: "admin",
		Map: RouteMap{
			"AdminGetStats": {
				Method:  "GET",
				Pattern: "stats",
				Handler: AdminGetStats,
			},
		},
	})
}

func AdminGetStats(req Request) (*Response, error) {
	_, err := AdminCheckAuth(req)
	if err != nil {
		return nil, err
	}
	return &Response{
		Data: map[string]interface{}{
			"locked_resource_count": resLock.CountLocked(),
		},
	}, nil
}
