package mdns

import (
	"fmt"
	"sync"

	"github.com/hashicorp/mdns"
)

type MDNS struct {
	entriesChan      chan *mdns.ServiceEntry
	detectedAdresses sync.Map
}

func New(service string) *MDNS {
	entriesChan := make(chan *mdns.ServiceEntry)

	mdns.Lookup(service, entriesChan)

	return &MDNS{entriesChan: entriesChan}
}

func (m *MDNS) Run() {
	for entry := range m.entriesChan {
		fmt.Printf("Got new entry: %v\n", entry)
		m.detectedAdresses.Store(entry.Host, entry.AddrV4.String())
	}
}

func (m *MDNS) Devices() map[string]string {
	devices := make(map[string]string)

	m.detectedAdresses.Range(func(key, value any) bool {
		keyString, ok := key.(string)
		if !ok {
			return false
		}

		valueString, ok := value.(string)
		if !ok {
			return false
		}

		devices[keyString] = valueString
		return true
	})

	return devices
}
