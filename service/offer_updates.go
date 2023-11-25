package service

import (
	"context"
	"github.com/cryptopunkscc/go-warpdrive"
	"log"
	"sort"
	"sync"
	"time"
)

type Notify func([]Notification)

type Notification struct {
	warpdrive.Peer
	warpdrive.Offer
	*warpdrive.Info
}

type offerUpdates struct {
	channel chan *offerService
	log     *log.Logger
	inMu    sync.Locker
	outMu   sync.Locker
	job     *sync.WaitGroup
	notify  Notify
}

func (c offerUpdates) Start(ctx context.Context) <-chan struct{} {
	receive := c.channel
	finish := make(chan struct{})
	go func() {
		buffer := newOfferUpdatesBuffer()
		defer func() {
			if buffer.len() > 0 {
				c.processUpdates(buffer)
			}
			close(finish)
		}()
		for {
			select {
			case update := <-receive:
				// Add received update to buffer
				status := update.OfferStatus
				buffer[status.In][status.Id] = update
			default:
				switch {
				case buffer.len() == 0:
					// There are no elements to proceed.
					// Wait for next element and continue buffer update.
					next := <-receive
					if next == nil {
						return
					}
					receive <- next
				default:
					// Start processing buffered elements:
					// Prepare array with sorted updates.
					startTime := time.Now().UnixNano()

					c.processUpdates(buffer)

					// Cleanup buffer
					buffer = newOfferUpdatesBuffer()
					// Measure work/sleep cycle to 1 second
					endTime := time.Now().UnixNano()
					workTime := endTime - startTime
					sleepTime := int64(time.Second) - workTime
					time.Sleep(time.Nanosecond * time.Duration(sleepTime))
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		c.job.Wait()
		time.Sleep(200 * time.Millisecond)
		close(receive)
		<-finish
		c.log.Println("finish updating offers")
	}()
	return finish
}

func (c offerUpdates) processUpdates(buffer offerUpdatesBuffer) {
	var updates []*offerService
	for _, b := range buffer {
		for _, next := range b {
			updates = append(updates, next)
		}
	}

	sort.Sort(byUpdateTime(updates))

	// Save updates in memory cache.
	c.inMu.Lock()
	c.outMu.Lock()
	for _, update := range updates {
		update.offerMemStorage.Save(*update.Offer)
	}
	c.inMu.Unlock()
	c.outMu.Unlock()

	// Save updates in storage.
	for _, update := range updates {
		if !update.IsOngoing() {
			update.offerFileStorage.Save(*update.Offer)
		}
	}

	// Notify listeners
	for _, update := range updates {
		update.statusBroadcast.Emit(update.OfferStatus)
		if update.Status == warpdrive.StatusAwaiting {
			update.offersBroadcast.Emit(*update.Offer)
		}
	}

	// Display system notification
	if n := c.notify; n != nil {
		arr := make([]Notification, len(updates))
		for i, update := range updates {
			arr[i] = update.notification()
		}
		n(arr)
	}
}

type offerUpdatesBuffer map[bool]map[warpdrive.OfferId]*offerService

func newOfferUpdatesBuffer() offerUpdatesBuffer {
	return offerUpdatesBuffer{true: {}, false: {}}
}

func (b offerUpdatesBuffer) len() int {
	return len(b[true]) + len(b[false])
}

type byUpdateTime []*offerService

func (a byUpdateTime) Len() int           { return len(a) }
func (a byUpdateTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byUpdateTime) Less(i, j int) bool { return a[i].Update < a[j].Update }

func (srv *offerService) notification() (n Notification) {
	o := *srv.Offer
	n = Notification{
		Peer:  *srv.peerStorage.Get()[o.Peer],
		Offer: o,
	}
	if o.IsOngoing() {
		n.Info = &o.Files[o.Index]
	}
	return
}
