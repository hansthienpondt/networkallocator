package main

import (
	"fmt"

	"github.com/hansthienpondt/networkallocator/pkg/asn"
	"github.com/hansthienpondt/networkallocator/pkg/vlan"
	"k8s.io/apimachinery/pkg/labels"
)

func main() {
	a := asn.NewASN(uint32(65000), labels.Set{})
	a2 := asn.NewASN(uint32(4294967295), labels.Set{})
	a3 := asn.NewASN(uint32(64500), labels.Set{})

	fmt.Println(a.IsPrivate(), a.IsDocumentation(), a.IsReserved())
	fmt.Println(a2.IsPrivate(), a2.IsDocumentation(), a2.IsReserved())
	fmt.Println(a3.IsPrivate(), a3.IsDocumentation(), a3.IsReserved())

	db := vlan.NewDB2()
	db[0] = labels.Set(map[string]string{"VLAN": "0", "status": "reserved", "custom": "value"})
	db[1] = labels.Set(map[string]string{"VLAN": "1", "status": "allocated"})
	db[2] = labels.Set(map[string]string{})
	db[4095] = labels.Set(map[string]string{"VLAN": "4095", "status": "reserved"})

	fmt.Println(db)
	db.Allocate(0)
	db.Allocate(2)
	db.Allocate(100)

	fmt.Println(db.Has(0), db.Has(2), db.Has(100), db.Delete(0), db)

	vdb := vlan.NewDB()
	v100, _ := vlan.NewVLAN(100, map[string]string{"key": "value"})
	v200, _ := vlan.NewVLAN(200, map[string]string{"key": "value"})

	vdb.Set(v100)
	vdb.Set(v200)
	fmt.Println(vdb.Has(100), vdb.Count())
	fmt.Println(vdb.Get(101))

	v, _ := vlan.NewVLAN(100, map[string]string{"key": "value"})
	fmt.Println(v, vdb)

	sel, _ := labels.Parse("status=reserved")
	fmt.Println(vdb.GetByLabel(sel))
	fmt.Println(vdb.GetAll())

}
