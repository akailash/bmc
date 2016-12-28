package main

import (
	"sort"
	"sync"
)

type keyrange struct {
	Llimit, Ulimit int64
}
type msgstore struct {
	sync.RWMutex
	ranges []keyrange
	keys   []int64
}

type wantedstore struct {
	sync.RWMutex
	cache map[int64][]string
}

func (store *msgstore) Add(m *NewMsg) {
	store.Lock()
	store.keys = append(store.keys, m.GetHead().GetMsgId())
	store.Unlock()
}
func (store *msgstore) GetKeys() (ranges []keyrange, keys []int64) {
	store.RLock()
	keys = make([]int64, len(store.keys))
	copy(keys, store.keys)
	ranges = make([]keyrange, len(store.ranges))
	copy(ranges, store.ranges)
	store.RUnlock()
	return ranges, keys
}

func (store *msgstore) DiffKeys(ranges []keyrange, keys []int64) (unknown []int64) {
	store.RLock()
	for _, r := range ranges {
		for i := r.Llimit; i <= r.Ulimit; i++ {
			if !CheckJsonDisk(i, StoreDir) {
				unknown = append(unknown, i)
			}
		}
	}
	for _, k := range keys {
		if !CheckJsonDisk(k, StoreDir) {
			unknown = append(unknown, k)
		}
	}

	store.RUnlock()
	return unknown
}
func removeRange(s []keyrange, i int) []keyrange {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

func (store *msgstore) Clean() {
	store.Lock()
	sort.Sort(int64arr(store.keys))
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
}

func (wanted *wantedstore) Add(keys []int64, host string) {
	wanted.Lock()
	for _, k := range keys {
		wanted.cache[k] = append(wanted.cache[k], host)
	}
	wanted.Unlock()

}
func (wanted *wantedstore) Delete(k int64) {
	wanted.Lock()
	delete(wanted.cache, k)
	wanted.Unlock()

}
