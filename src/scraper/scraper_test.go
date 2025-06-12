package scraper

import (
	"fmt"
	"testing"
)

func Test_fetchAll(t *testing.T) {
	offerts := FetchAll()
	pageMin := 20
	pageMax := 30
	perPage := 50
	offertMin := perPage * pageMin
	offertMax := perPage * pageMax

	oto := 0
	for _, o := range offerts {
		if o.Type == "otodom" {
			fmt.Printf("Otodom offert %v", o)
			oto += 1
		}
	}

	fmt.Printf("Otodoms offerts %d", oto)

	if len(offerts) < offertMin || len(offerts) > offertMax {
		t.Errorf("the amount of offerts is %d when expected between %d, %d", len(offerts), offertMin, offertMax)
	}
}

func Test_fetchPage(t *testing.T) {
}
