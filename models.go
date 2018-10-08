// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import (
	"sync"
)

const (
	odCascade             = "cascade"
	odSetNULL             = "set_null"
	odSetDefault          = "set_default"
	odDoNothing           = "do_nothing"
	defaultStructTagName  = "orm"
	defaultStructTagDelim = ";"
)

var (
	modelCache = &_modelCache{
		cache:           make(map[string]*modelInfo),
		cacheByFullName: make(map[string]*modelInfo),
	}
)

// model info collection
type _modelCache struct {
	sync.RWMutex    // only used outsite for bootStrap
	orders          []string
	cache           map[string]*modelInfo
	cacheByFullName map[string]*modelInfo
	done            bool
}

// get all model info
func (mc *_modelCache) all() map[string]*modelInfo {
	mc.RLock()
	defer mc.RUnlock()
	m := make(map[string]*modelInfo, len(mc.cache))
	for k, v := range mc.cache {
		m[k] = v
	}
	return m
}

// get orderd model info
func (mc *_modelCache) allOrdered() []*modelInfo {
	mc.RLock()
	defer mc.RUnlock()
	m := make([]*modelInfo, 0, len(mc.orders))
	for _, table := range mc.orders {
		m = append(m, mc.cache[table])
	}
	return m
}

// get model info by table name
func (mc *_modelCache) get(table string) (mi *modelInfo, ok bool) {
	mc.RLock()
	defer mc.RUnlock()
	mi, ok = mc.cache[table]
	return
}

// get model info by full name
func (mc *_modelCache) getByFullName(name string) (mi *modelInfo, ok bool) {
	mc.RLock()
	defer mc.RUnlock()
	mi, ok = mc.cacheByFullName[name]
	return
}

// set model info to collection
func (mc *_modelCache) set(table string, mi *modelInfo) *modelInfo {
	mc.Lock()
	defer mc.Unlock()
	mii := mc.cache[table]
	mc.cache[table] = mi
	mc.cacheByFullName[mi.fullName] = mi
	if mii == nil {
		mc.orders = append(mc.orders, table)
	}
	return mii
}

func (mc *_modelCache) removeByFullName(name string) *modelInfo {
	mc.Lock()
	defer mc.Unlock()
	mi, ok := mc.cacheByFullName[name]
	if ok {
		delete(mc.cacheByFullName, name)
		delete(mc.cache, mi.table)
		if mc.orders != nil && len(mc.orders) > 0 {
			orders := make([]string, 0, len(mc.orders)-1)
			for _, order := range mc.orders {
				if order != mi.table {
					orders = append(orders, order)
				}
			}
			mc.orders = orders
		}
	}
	return mi
}

func (mc *_modelCache) remove(table string) *modelInfo {
	mc.Lock()
	defer mc.Unlock()
	mi, ok := mc.cache[table]
	if ok {
		delete(mc.cache, table)
		delete(mc.cacheByFullName, mi.fullName)
		if mc.orders != nil && len(mc.orders) > 0 {
			orders := make([]string, 0, len(mc.orders)-1)
			for _, order := range mc.orders {
				if order != table {
					orders = append(orders, order)
				}
			}
			mc.orders = orders
		}
	}
	return mi
}

// clean all model info.
func (mc *_modelCache) clean() {
	mc.Lock()
	defer mc.Unlock()
	mc.orders = make([]string, 0)
	mc.cache = make(map[string]*modelInfo)
	mc.cacheByFullName = make(map[string]*modelInfo)
	mc.done = false
}

// ResetModelCache Clean model cache. Then you can re-RegisterModel.
// Common use this api for test case.
func ResetModelCache() {
	modelCache.clean()
}
