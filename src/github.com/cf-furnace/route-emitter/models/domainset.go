package models

type DomainSet map[string]struct{}

func NewDomainSet(items ...string) DomainSet {
	d := DomainSet{}
	for _, i := range items {
		d[i] = struct{}{}
	}
	return d
}
