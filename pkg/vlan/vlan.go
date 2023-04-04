package vlan

import (
	"fmt"

	"github.com/mdlayher/ethernet"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	MaxVLAN = uint16(4096)
)

type VLAN struct {
	ethernet.VLAN
	labels labels.Set
}
type VLANs []VLAN

func (vs VLANs) SetLabels(l labels.Set) VLANs {
	for k, v := range vs {
		vs[k] = v.SetLabels(l)
	}
	return vs
}

func (v VLAN) ID() uint16                  { return v.VLAN.ID }
func (v VLAN) Labels() labels.Set          { return v.labels }
func (v VLAN) String() string              { return fmt.Sprintf("%d %s", v.ID(), v.Labels().String()) }
func (v VLAN) SetLabels(l labels.Set) VLAN { v.labels = l; return v }

func NewVLAN(id uint16, l map[string]string) (VLAN, error) {
	var label labels.Set

	if l == nil {
		label = labels.Set{}
	} else {
		label = labels.Set(l)
	}
	if id > ethernet.VLANMax {
		return VLAN{VLAN: ethernet.VLAN{ID: id}}, ethernet.ErrInvalidVLAN
	}
	return VLAN{
		VLAN:   ethernet.VLAN{ID: id},
		labels: label,
	}, nil
}
