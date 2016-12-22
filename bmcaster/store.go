package main

import (
	"sort"
	"sync"
)

type keyrange struct {
	Llimit, Ulimit int
}
type msgstore struct {
	sync.RWMutex
	cache  map[int]Msg
	ranges []keyrange
	keys   []int
}

type wantedstore struct {
	sync.RWMutex
	cache map[int][]string
}

func (store *msgstore) Add(m Msg) {
	store.Lock()
	store.cache[m.MsgID] = m
	store.keys = append(store.keys, m.MsgID)
	store.Unlock()
}
func (store *msgstore) GetKeys() (ranges []keyrange, keys []int) {
	store.RLock()
	keys = make([]int, len(store.keys))
	copy(keys, store.keys)
	ranges = make([]keyrange, len(store.ranges))
	copy(ranges, store.ranges)
	store.RUnlock()
	return ranges, keys
}

func (store *msgstore) DiffKeys(ranges []keyrange, keys []int) (unknown []int) {
	store.RLock()
	for _, r := range ranges {
		for i := r.Llimit; i <= r.Ulimit; i++ {
			if _, found := store.cache[i]; !found {
				if !CheckJsonDisk(i, StoreDir) {
					unknown = append(unknown, i)
				}
			}

		}
	}
	for _, k := range keys {
		if _, found := store.cache[k]; !found {
			if !CheckJsonDisk(k, StoreDir) {
				unknown = append(unknown, k)
			}
		}
	}

	store.RUnlock()
	return unknown
}
func removeRange(s []keyrange, i int) []keyrange {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
func (store *msgstore) Clean() {
	store.Lock()
	bin := store.cache
	store.cache = make(map[int]Msg)
	sort.Ints(store.keys)
	for _, k := range store.keys {
		added := false
		for i, r := range store.ranges {
			if k == r.Llimit-1 {
				store.ranges[i].Llimit = k
				added = true
				break
			} else if k == r.Ulimit+1 {
				store.ranges[i].Ulimit = k
				added = true
				break
			} else if k >= r.Llimit && k <= r.Ulimit {
				added = true
				break
			}
			if i > 1 && store.ranges[i-1].Ulimit == store.ranges[i].Llimit {
				store.ranges[i-1].Ulimit = store.ranges[i].Ulimit
				store.ranges = removeRange(store.ranges, i)
			}
		}
		if added == false {
			//Create new range for this value
			store.ranges = append(store.ranges, keyrange{k, k})
		}
	}
	//This lets Go GC collect the memory from the slice
	store.keys = nil
	store.Unlock()
	for k := range bin {
		//For a map, each element should be deleted
		delete(bin, k)
	}
}

func (wanted *wantedstore) Add(keys []int, host string) {
	wanted.Lock()
	for _, k := range keys {
		wanted.cache[k] = append(wanted.cache[k], host)
	}
	wanted.Unlock()

}
func (wanted *wantedstore) Delete(k int) {
	wanted.Lock()
	delete(wanted.cache, k)
	wanted.Unlock()

}
