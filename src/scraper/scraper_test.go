package scraper

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"house-pricer/types"
)

func Test_fetchAll(t *testing.T) {
	resultChannel := make(chan []types.Offert)
	FetchAll(resultChannel)
	offerts := []types.Offert{}

	for batch := range resultChannel {
		offerts = append(offerts, batch...)
	}

	pageMin := 20
	pageMax := 30
	perPage := 50
	offertMin := perPage * pageMin
	offertMax := perPage * pageMax

	oto := 0
	for _, o := range offerts {
		if o.Type == "otodom" {
			fmt.Printf("Otodom offert %v\n", o)
			oto += 1
		}
	}

	fmt.Printf("Otodoms offerts %d\n", oto)

	if len(offerts) < offertMin || len(offerts) > offertMax {
		t.Errorf("the amount of offerts is %d when expected between %d, %d", len(offerts), offertMin, offertMax)
	}
}

// func Test_fetchAllMemoryUse(t *testing.T) {
// 	done := make(chan struct{})
// 	defer close(done)
// 	go monitorMemory(done)
//
// 	resultChannel := make(chan []types.Offert)
// 	FetchAll(resultChannel)
//
// 	for batch := range resultChannel {
// 		fmt.Println(batch)
// 	}
//
// 	time.Sleep(200 * time.Millisecond)
// }

func Test_fetchPage(t *testing.T) {
}

func monitorMemory(done <-chan struct{}) {
	var peak uint64
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.Alloc > peak {
				peak = m.Alloc
			}
		case <-done:
			fmt.Printf("Peak memory usage: %.2f MB\n", float64(peak)/1024/1024)
			return
		}
	}
}
