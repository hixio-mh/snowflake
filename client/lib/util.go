package lib

import (
	"log"
	"time"
)

const (
	LogTimeInterval = 5
)

type BytesLogger interface {
	Log()
	AddOutbound(int)
	AddInbound(int)
}

// Default BytesLogger does nothing.
type BytesNullLogger struct{}

func (b BytesNullLogger) Log()                   {}
func (b BytesNullLogger) AddOutbound(amount int) {}
func (b BytesNullLogger) AddInbound(amount int)  {}

// BytesSyncLogger uses channels to safely log from multiple sources with output
// occuring at reasonable intervals.
type BytesSyncLogger struct {
	OutboundChan chan int
	InboundChan  chan int
	Outbound     int
	Inbound      int
	OutEvents    int
	InEvents     int
}

func (b *BytesSyncLogger) Log() {
	var amount int
	output := func() {
		log.Printf("Traffic Bytes (in|out): %d | %d -- (%d OnMessages, %d Sends)",
			b.Inbound, b.Outbound, b.InEvents, b.OutEvents)
		b.Outbound = 0
		b.OutEvents = 0
		b.Inbound = 0
		b.InEvents = 0
	}
	last := time.Now()
	for {
		select {
		case amount = <-b.OutboundChan:
			b.Outbound += amount
			b.OutEvents++
			if time.Since(last) > time.Second*LogTimeInterval {
				last = time.Now()
				output()
			}
		case amount = <-b.InboundChan:
			b.Inbound += amount
			b.InEvents++
			if time.Since(last) > time.Second*LogTimeInterval {
				last = time.Now()
				output()
			}
		case <-time.After(time.Second * LogTimeInterval):
			if b.InEvents > 0 || b.OutEvents > 0 {
				output()
			}
		}
	}
}

func (b *BytesSyncLogger) AddOutbound(amount int) {
	b.OutboundChan <- amount
}

func (b *BytesSyncLogger) AddInbound(amount int) {
	b.InboundChan <- amount
}
