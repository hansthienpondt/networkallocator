package asn

import (
	"sync"

	"k8s.io/apimachinery/pkg/labels"
)

type ASN struct {
	mu     *sync.RWMutex
	asn    uint32
	labels labels.Set
}

func NewASN(asn uint32, labels labels.Set) ASN {
	return ASN{
		mu:     &sync.RWMutex{},
		asn:    asn,
		labels: labels,
	}
}
func (a ASN) IsDocumentation() bool {
	// RFC 5398
	// return true when ASN is a documentation ASN, false by default.
	// 16-bit (2-byte) Private ASN
	if a.asn >= 64496 && a.asn <= 64511 {
		return true
	}
	return false
}
func (a ASN) IsPrivate() bool {
	// RFC 6996
	// return true when ASN is private, false by default.
	// 16-bit (2-byte) Private ASN
	if a.asn >= 64512 && a.asn <= 65534 {
		return true
	}
	// 32-bit (4-byte) Private ASN
	if a.asn >= 4200000000 && a.asn <= 4294967294 {
		return true
	}
	return false
}

func (a ASN) IsReserved() bool {
	// RFC7300
	// return true when ASN is reserved by IANA.
	if a.asn == 65535 || a.asn == 4294967295 {
		return true
	}
	return false
}
