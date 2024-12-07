package ivy

import (
	"sync"
)

type ManagerRecord struct {
	pageNum int
	copySet []int
	owner   int
}

type NodeRecord struct {
	pageNum     int
	accessRight int
}

type ManagerRecords struct {
	lock    sync.RWMutex
	Records []ManagerRecord
}

type NodeRecords struct {
	lock    sync.RWMutex
	Records []NodeRecord
}

type Page struct {
	rwLock  sync.RWMutex
	content string
}

func (mrs *ManagerRecords) contains(pageNum int) bool {
	mrs.lock.RLock()
	defer mrs.lock.RUnlock()
	for _, record := range mrs.Records {
		if record.pageNum == pageNum {
			return true
		}
	}
	return false
}

func (nrs *NodeRecords) contains(pageNum int) bool {
	nrs.lock.RLock()
	defer nrs.lock.RUnlock()
	for _, record := range nrs.Records {
		if record.pageNum == pageNum {
			return true
		}
	}
	return false
}
