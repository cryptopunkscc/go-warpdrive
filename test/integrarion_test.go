package test

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/go-apphost-jrpc/android/notify"
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/cryptopunkscc/go-warpdrive/android"
	"github.com/cryptopunkscc/go-warpdrive/proto"
	"github.com/cryptopunkscc/go-warpdrive/start"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestMain_(t *testing.T) {
	// Start test notification service
	cancelFunc := notify.TestServer(t, false)
	defer cancelFunc()

	time.Sleep(600)

	ctx, shutdown := context.WithCancel(context.Background())

	defer func() {
		shutdown()
		if err := os.RemoveAll("test_dir"); err != nil {
			log.Panicln(err)
		}
	}()

	// Start warpdrive service
	go func(t *testing.T) {
		client := notify.NewTestClient()
		defer func() {
			if err := client.Close(); err != nil {
				t.Error(err)
			}
		}()
		if err := start.Warpdrive(ctx, start.Args{
			Logger:       log.Default(),
			Cache:        "test_dir",
			Store:        "test_dir",
			CreateNotify: android.CreateNotify(client),
		}); err != nil {
			t.Error(err)
		}
	}(t)

	time.Sleep(time.Second)

	t.Run("main flow", func(t *testing.T) {
		// Init offer subscription
		offerClient, cancel := testClient(t)
		defer cancel()
		go func() {
			offers, err := offerClient.ListenOffers(warpdrive.FilterAll)
			if err != nil {
				t.Error(err)
			}
			for offer := range offers {
				log.Println("offer:", offer)
			}
		}()

		// Init status subscription
		statusClient, cancel := testClient(t)
		defer cancel()
		go func() {
			offers, err := statusClient.ListenStatus(warpdrive.FilterAll)
			if err != nil {
				t.Error(err)
			}
			for offer := range offers {
				log.Println("status: ", offer)
			}
		}()

		// Execute main flow
		mainClient, cancel := testClient(t)
		defer cancel()

		offer, err := mainClient.CreateOffer("", "integrarion_test.go")
		if err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, false, offer.In)
			assert.Equal(t, "awaiting", offer.Status)
		}

		time.Sleep(time.Millisecond * 600)

		if offers, err := mainClient.ListOffers(warpdrive.FilterAll); err != nil {
			t.Error(err)
		} else {
			for _, o := range offers {
				log.Println(o)
			}
		}

		time.Sleep(time.Millisecond * 600)

		err = mainClient.AcceptOffer(offer.Id)
		if err != nil {
			t.Error(err)
		}

		time.Sleep(time.Millisecond * 600)
	})
}

func testClient(t *testing.T) (api warpdrive.LocalApi, cancel func()) {
	c := proto.NewClient()
	c, err := c.Connect(id.Identity{}, warpdrive.Port)
	if err != nil {
		t.Fatal(err)
		return
	}
	api = c
	cancel = func() {
		if err = c.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return
}
