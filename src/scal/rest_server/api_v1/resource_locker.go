package api_v1

import (
	"sort"
	"sync"
	"scal/settings"
)

const (
	restype_user = iota
	restype_user_login
	restype_event
	restype_group
	restype_COUNT
)

var resLock = NewResourceLocker()

func NewResourceLocker() *ResourceLocker {
	rl := &ResourceLocker{}
	for i := 0; i < restype_COUNT; i++ {
		rl.mutexes[i] = &sync.RWMutex{}
		rl.lockedMaps[i] = map[string]bool{}
	}
	return rl
}

type ResourceLocker struct {
	mutexes    [restype_COUNT]*sync.RWMutex
	lockedMaps [restype_COUNT]map[string]bool
}

func (rl *ResourceLocker) CountLocked() map[string]int {
	return map[string]int{
		"user":       rl.CountLockedResource(restype_user),
		"user_login": rl.CountLockedResource(restype_user_login),
		"event":      rl.CountLockedResource(restype_event),
		"group":      rl.CountLockedResource(restype_group),
	}
}

func (rl *ResourceLocker) CountLockedResource(resType int) int {
	mutex := rl.mutexes[resType]
	mutex.RLock()
	defer mutex.RUnlock()
	return len(rl.lockedMaps[resType])
}

func (rl *ResourceLocker) ListLocked() map[string][]string {
	data := map[string][]string{
		"user":       rl.ListLockedResource(restype_user),
		"user_login": rl.ListLockedResource(restype_user_login),
		"event":      rl.ListLockedResource(restype_event),
		"group":      rl.ListLockedResource(restype_group),
	}
	for _, ids := range data {
		sort.Strings(ids)
	}
	return data
}

// order is random
func (rl *ResourceLocker) ListLockedResource(resType int) []string {
	mutex := rl.mutexes[resType]
	lockedMap := rl.lockedMaps[resType]
	mutex.RLock()
	defer mutex.RUnlock()
	ids := make([]string, 0, len(lockedMap))
	for id, locked := range lockedMap {
		if locked {
			ids = append(ids, id)
		}
	}
	return ids
}

// returns failed, unlock
func (rl *ResourceLocker) Resource(resType int, resId string) (bool, func()) {
	mutex := rl.mutexes[resType]
	lockedMap := rl.lockedMaps[resType]
	mutex.Lock()
	defer mutex.Unlock()
	if lockedMap[resId] {
		return true, nil
	}
	if settings.RESOURCE_LOCK_REDIS_ENABLE {
		// TODO: redis lock
	}
	lockedMap[resId] = true
	return false, func() {
		mutex.Lock()
		defer mutex.Unlock()
		delete(lockedMap, resId)
	}
}

// returns failed, unlock
func (rl *ResourceLocker) User(email string) (bool, func()) {
	return rl.Resource(restype_user, email)
}

// returns failed, unlock
func (rl *ResourceLocker) UserLogin(email string, ip string) (bool, func()) {
	return rl.Resource(restype_user_login, email+"-"+ip)
}

// returns failed, unlock
func (rl *ResourceLocker) Event(eventId string) (bool, func()) {
	return rl.Resource(restype_event, eventId)
}

// returns failed, unlock
func (rl *ResourceLocker) Group(groupId string) (bool, func()) {
	return rl.Resource(restype_group, groupId)
}
