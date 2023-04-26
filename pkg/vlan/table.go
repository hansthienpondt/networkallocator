package vlan

import (
	"fmt"
	"sort"
	"sync"

	"github.com/mdlayher/ethernet"
	"k8s.io/apimachinery/pkg/labels"
)

type DB struct {
	mu     *sync.RWMutex
	store  map[uint16]VLAN
	labels labels.Set
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

func (it *DBIterator) HasConsecutive(i uint16) bool {
	// i includes current value, deduct the current value.
	i--
	if it.current+int(i) >= len(it.keys) {
		return false
	}
	return it.keys[it.current]+int(i) == it.keys[it.current+int(i)]
}

func NewDB() *DB {
	db := &DB{
		mu:     &sync.RWMutex{},
		store:  make(map[uint16]VLAN),
		labels: labels.Set{},
	}
	vlan0, _ := NewVLAN(0, map[string]string{"type": "untagged", "status": "reserved"})
	vlan1, _ := NewVLAN(1, map[string]string{"type": "default", "status": "reserved"})
	vlan4095, _ := NewVLAN(4095, map[string]string{"type": "reserved", "status": "reserved"})
	db.store[vlan0.ID()] = vlan0
	db.store[vlan1.ID()] = vlan1
	db.store[vlan4095.ID()] = vlan4095
	return db
}

func (db *DB) Add(vlan VLAN) error {
	if db.Has(vlan.ID()) {
		return fmt.Errorf("VLAN %d already exists in the VLAN database ", vlan.ID())
	}
	return db.Set(vlan)
}

func (db *DB) AddVlanList(vlans VLANs) error {
	var err error
	for _, vlan := range vlans {
		err = db.Add(vlan)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) FindAllocateVlan(l labels.Set) (VLAN, error) {
	var err error
	var iter *DBIterator = db.IterateFree()
	var vlan VLAN

	for iter.Next() {
		vlan = iter.Value()
		vlan = vlan.SetLabels(l)

		err = db.Add(vlan)
		// if no error has been returned, adding the vlan was successful.
		if err == nil {
			return vlan, err
		}
	}
	return VLAN{}, fmt.Errorf("Could find/allocate a new VLAN")
}
func (db *DB) FindVlanRange(min, amount uint16) VLANs {
	/*
		min specifies the minimum boundary from which available Vlan Ranges should be returned.
		amount specifies the number of consecutive vlans required
	*/
	var vlans VLANs
	var free *DBIterator = db.IterateFree()

	for free.Next() {
		if free.Value().ID() < min {
			continue
		}
		if free.HasConsecutive(amount) {
			min = free.Value().ID()
			vlans = append(vlans, free.Value())
			amount--
		}
		if amount == 0 {
			return vlans
		}
	}

	return make(VLANs, 0)

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
		db.mu.Lock()
		defer db.mu.Unlock()

		db.store[vlan.ID()] = vlan
		return nil
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
		if selector.Matches(iter.Value().Labels) {
			vlans = append(vlans, iter.Value())
		}
	}
	return vlans
}
func (db *DB) Labels() labels.Set {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.labels
}
func (db *DB) SetLabels(l labels.Set) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.labels = l
}
func (db *DB) MergeLabels(l labels.Set) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.labels = labels.Merge(db.labels, l)
}
func (db *DB) DeleteLabels() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.labels = labels.Set{}
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
	db.mu.RLock()
	defer db.mu.RUnlock()

	_, ok := db.store[id]
	return ok
}

func (db *DB) Delete(id uint16) error {
	switch id {
	case 0:
		return fmt.Errorf("VLAN %d is the untagged VLAN, cannot be deleted from the database", id)
	case 1:
		return fmt.Errorf("VLAN %d is the default VLAN, cannot be deleted from the database", id)
	case 4095:
		return fmt.Errorf("VLAN %d is reserved, cannot be deleted from the database", id)
	default:
		db.mu.Lock()
		defer db.mu.Unlock()

		delete(db.store, id)
		return nil
	}
}

func (db *DB) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return len(db.store)
}
func (db *DB) CountFree() int {
	return ethernet.VLANMax - db.Count() + 1
}

func (db *DB) Iterate() *DBIterator {
	var vlans []int
	db.mu.RLock()
	defer db.mu.RUnlock()
	for k := range db.store {
		vlans = append(vlans, int(k))
	}
	sort.Ints(vlans)

	return &DBIterator{current: -1, keys: vlans, db: db.store}
}

func (db *DB) IterateFree() *DBIterator {
	var vlans []int
	var store map[uint16]VLAN = make(map[uint16]VLAN)

	db.mu.RLock()
	defer db.mu.RUnlock()

	for id := 0; id < ethernet.VLANMax; id++ {
		_, exists := db.store[uint16(id)]
		if !exists {
			vlans = append(vlans, id)
			store[uint16(id)] = VLAN{
				VLAN:   ethernet.VLAN{ID: uint16(id)},
				Labels: labels.Set{},
			}
		}
	}
	sort.Ints(vlans)

	return &DBIterator{current: -1, keys: vlans, db: store}

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
