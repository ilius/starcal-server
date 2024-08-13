package api_v1

import (
	"github.com/ilius/ripo"
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
			"AdminListLockedResources": {
				Method:  "GET",
				Pattern: "locked-resources",
				Handler: AdminListLockedResources,
			},
		},
	})
}

func AdminGetStats(req ripo.Request) (*ripo.Response, error) {
	_, err := AdminCheckAuth(req)
	if err != nil {
		return nil, err
	}
	return &ripo.Response{
		Data: map[string]any{
			"locked_resource_count": resLock.CountLocked(),
		},
	}, nil
}

func AdminListLockedResources(req ripo.Request) (*ripo.Response, error) {
	_, err := AdminCheckAuth(req)
	if err != nil {
		return nil, err
	}
	return &ripo.Response{
		Data: resLock.ListLocked(),
	}, nil
}
