package api_v1

import (
	"gopkg.in/mgo.v2/bson"
	"sync"
)

const (
	restype_user  int = 0
	restype_event int = 1
	restype_group int = 2
)

var resLock = &ResourceLocker{
	mutexes:    [3]sync.RWMutex{},
	lockedMaps: [3]map[string]bool{},
}

type ResourceLocker struct {
	mutexes    [3]sync.RWMutex
	lockedMaps [3]map[string]bool
}

func (rl *ResourceLocker) CountLocked() (counts [3]int) {
	for resType:=0; resType<3; resType++ {
		counts[resType] = rl.CountLockedResource(resType)
	}
	return
}

func (rl *ResourceLocker) CountLockedResource(resType int) (int) {
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
