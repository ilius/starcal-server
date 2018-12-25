package api_v1

import (
	"sync"

	"gopkg.in/mgo.v2/bson"
)

const (
	restype_user  int = 0
	restype_event int = 1
	restype_group int = 2
)

var resLock = NewResourceLocker()

func NewResourceLocker() *ResourceLocker {
	rl := &ResourceLocker{}
	for i := 0; i < len(rl.lockedMaps); i++ {
		rl.lockedMaps[i] = map[string]bool{}
	}
	return rl
}

type ResourceLocker struct {
	mutexes    [3]sync.RWMutex
	lockedMaps [3]map[string]bool
}

func (rl *ResourceLocker) CountLocked() map[string]int {
	return map[string]int{
		"user":  rl.CountLockedResource(restype_user),
		"event": rl.CountLockedResource(restype_event),
		"group": rl.CountLockedResource(restype_group),
	}
}

func (rl *ResourceLocker) CountLockedResource(resType int) int {
	mutex := rl.mutexes[resType]
	mutex.RLock()
	defer mutex.RUnlock()
	return len(rl.lockedMaps[resType])
}

// returns failed, unlock
func (rl *ResourceLocker) Resource(resType int, resId string) (bool, func()) {
	mutex := rl.mutexes[resType]
	lockedMaps := rl.lockedMaps[resType]
	mutex.Lock()
	defer mutex.Unlock()
	if lockedMaps[resId] {
		return true, nil
	}
	lockedMaps[resId] = true
	return false, func() {
		mutex.Lock()
		defer mutex.Unlock()
		delete(lockedMaps, resId)
	}
}

// returns failed, unlock
func (rl *ResourceLocker) User(email string) (bool, func()) {
	return rl.Resource(restype_user, email)
}

// returns failed, unlock
func (rl *ResourceLocker) Event(eventId bson.ObjectId) (bool, func()) {
	return rl.Resource(restype_event, eventId.Hex())
}

// returns failed, unlock
func (rl *ResourceLocker) Group(groupId bson.ObjectId) (bool, func()) {
	return rl.Resource(restype_group, groupId.Hex())
}