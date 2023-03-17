package vlan

import (
	"fmt"
	"sort"
	"sync"

	"github.com/mdlayher/ethernet"
	"k8s.io/apimachinery/pkg/labels"
)

type DB struct {
	mu    *sync.RWMutex
	store map[uint16]VLAN
}

type DBIterator struct {
	current int
	keys    []int
	db      map[uint16]VLAN
}

func (it *DBIterator) Value() VLAN {
	return it.db[uint16(it.keys[it.current])]
}

func (it *DBIterator) Next() bool {
	it.current++

	return it.current < len(it.keys)
}

func NewDB() *DB {
	db := &DB{
		mu:    &sync.RWMutex{},
		store: make(map[uint16]VLAN),
	}
	vlan0, _ := NewVLAN(0, map[string]string{"type": "untagged", "status": "reserved"})
	vlan1, _ := NewVLAN(1, map[string]string{"type": "default", "status": "reserved"})
	vlan4095, _ := NewVLAN(4095, map[string]string{"type": "reserved", "status": "reserved"})
	db.add(vlan0)
	db.add(vlan1)
	db.add(vlan4095)
	return db
}

func (db *DB) add(vlan VLAN) error {
	db.store[vlan.ID()] = vlan
	return nil
}

func (db *DB) Set(vlan VLAN) error {
	switch vlan.ID() {
	case 0:
		return fmt.Errorf("VLAN %d is the untagged VLAN, cannot be added to the database", vlan.ID())
	case 1:
		return fmt.Errorf("VLAN %d is the default VLAN, cannot be added to the database", vlan.ID())
	case 4095:
		return fmt.Errorf("VLAN %d is reserved, cannot be added to the database", vlan.ID())
	default:
		return db.add(vlan)
	}
}

func (db *DB) Get(id uint16) (VLAN, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	vlan, ok := db.store[id]
	if !ok {
		return VLAN{VLAN: ethernet.VLAN{ID: id}}, fmt.Errorf("no match found")
	}
	return vlan, nil

}
func (db *DB) GetByLabel(selector labels.Selector) VLANs {
	var vlans VLANs
	db.mu.RLock()
	defer db.mu.RUnlock()

	iter := db.Iterate()
	for iter.Next() {
		if selector.Matches(iter.Value().Labels()) {
			vlans = append(vlans, iter.Value())
		}
	}
	return vlans
}

func (db *DB) GetAll() VLANs {
	var vlans VLANs
	db.mu.RLock()
	defer db.mu.RUnlock()

	iter := db.Iterate()
	for iter.Next() {
		vlans = append(vlans, iter.Value())
	}
	return vlans
}

func (db *DB) Has(id uint16) bool {
	_, ok := db.store[id]
	return ok
}

func (db *DB) Delete(id uint16) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	switch id {
	case 0:
		return fmt.Errorf("VLAN %d is the untagged VLAN, cannot be deleted from the database", id)
	case 1:
		return fmt.Errorf("VLAN %d is the default VLAN, cannot be deleted from the database", id)
	case 4095:
		return fmt.Errorf("VLAN %d is reserved, cannot be deleted from the database", id)
	default:
		delete(db.store, id)
		return nil
	}
}

func (db *DB) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.store)
}

func (db *DB) Iterate() *DBIterator {
	var vlans []int
	for k := range db.store {
		vlans = append(vlans, int(k))
	}
	sort.Ints(vlans)

	return &DBIterator{current: -1, keys: vlans, db: db.store}
}

// Alternative Implementation

type DB2 [MaxVLAN]labels.Set

func (d *DB2) Allocate(id uint16) {
	fmt.Println(len(d[id]))
}
func (d *DB2) Has(e uint16) bool {

	return len(d[e]) > 0
}
func (d *DB2) Delete(e uint16) bool {
	if len(d[e]) > 0 {
		d[e] = labels.Set{}
		return true
	}
	return false
}

func NewDB2() *DB2 {
	return &DB2{}
}
