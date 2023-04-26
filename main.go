package main

import (
	"encoding/json"
	"fmt"

	. "github.com/hansthienpondt/networkallocator/pkg/vlan"
	"k8s.io/apimachinery/pkg/labels"
)

func main() {

	db := NewDB()
	db.MergeLabels(map[string]string{"type": "router"})

	v100, _ := NewVLAN(100, map[string]string{"key": "value"})
	v200, _ := NewVLAN(200, map[string]string{"key": "value"})

	db.Set(v100)
	db.Set(v200)
	fmt.Println(db.Has(100), db.Count())
	fmt.Println(db.Get(101))

	sel, _ := labels.Parse("status=reserved")
	fmt.Println(db.GetByLabel(sel))

	gap, _ := NewVLAN(1005, map[string]string{"created": "gap"})
	db.Add(gap)
	l := db.FindVlanRange(995, 10).SetLabels(map[string]string{"type": "range"})

	db.AddVlanList(l)

	fmt.Println(db.GetAll())
	/*
	   iter := db.IterateFree()

	   	for iter.Next() {
	   		fmt.Println(iter.Value(), iter.HasConsecutive(10))
	   	}
	*/
	j, _ := json.Marshal(gap.Labels)
	k := labels.Set{}

	json.Unmarshal(j, &k)

	fmt.Println(string(j), k.String())
	j2v, _ := json.Marshal(gap)
	j2k, _ := gap.MarshalBinary()
	fmt.Println(j2k, j2v)

	vlan, err := db.FindAllocateVlan(map[string]string{"created": "FindAllocateVlan"})

	fmt.Println(vlan, err)
	fmt.Println(db.GetAll())

}
