package main

import (
	"log"
	"runtime"
	"sort"
	"sync"
)

type keyrange struct {
	Llimit, Ulimit int64
}
type msgstore struct {
	sync.RWMutex
	ranges []keyrange
}

type keyranges []keyrange

func (slice keyranges) Len() int {
	return len(slice)
}

func (slice keyranges) Less(i, j int) bool {
	return slice[i].Ulimit <= slice[j].Llimit
}

func (slice keyranges) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func (store *msgstore) Add(k int64) {
	var added bool = false
	store.Lock()
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
			log.Println(k, "added again to ", r)
			break
		}

	}
	if added == false {
		//Create new range for this value
		store.ranges = append(store.ranges, keyrange{k, k})
	}
	sort.Sort(keyranges(store.ranges))
	for i, _ := range store.ranges {
		if (i > 0) && (store.ranges[i-1].Ulimit == (store.ranges[i].Llimit - 1)) {
			store.ranges[i].Llimit = store.ranges[i-1].Llimit
			store.ranges = removeRange(store.ranges, i-1)
			break
		}
	}
	store.Unlock()
}
func (store *msgstore) GetKeys() (ranges []keyrange) {
	store.RLock()
	ranges = make([]keyrange, len(store.ranges))
	copy(ranges, store.ranges)
	store.RUnlock()
	return ranges
}
func (store *msgstore) IsEmpty() bool {
	return len(store.ranges) == 0
}
func (store *msgstore) DiffKeys(ranges []keyrange) (unknown []int64) {
	store.RLock()
	log.Println("DiffKeys()", store.ranges, ranges)
	for index, r := range ranges {
		var matched bool = false
		for _, sr := range store.ranges {
			if r.Llimit >= sr.Llimit {
				if r.Ulimit <= sr.Ulimit {
					matched = true
				} else {
					if r.Llimit <= sr.Ulimit {
						ranges[index].Llimit = sr.Ulimit + 1
					}
				}
			} else {
				if r.Ulimit >= sr.Llimit {
					if r.Ulimit > sr.Ulimit {
						//add earlier unknown keys and modify range
						for i := ranges[index].Llimit; i < sr.Llimit; i++ {
							unknown = append(unknown, i)
						}
						ranges[index].Llimit = sr.Ulimit + 1
					} else {
						ranges[index].Ulimit = sr.Llimit - 1
					}
				}
			}
		}
		if matched == true {
			continue
		}
		for i := ranges[index].Llimit; i <= ranges[index].Ulimit; i++ {
			unknown = append(unknown, i)
		}
	}
	store.RUnlock()
	log.Println("DiffKeys(): unknown", unknown)
	return unknown
}
func removeRange(s []keyrange, i int) []keyrange {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (store *msgstore) Clean() {
	//runtime.Gosched()
	//runtime.GC()
	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//log.Printf("%+v\n", m)
	log.Println("NumGoroutine", runtime.NumGoroutine())
}
