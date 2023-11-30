package android

import (
	"github.com/cryptopunkscc/go-apphost-jrpc/android/notify"
	"github.com/cryptopunkscc/go-warpdrive"
	"testing"
	"time"
)

func TestClient_All(t *testing.T) {
	cancelFunc := notify.TestServer(t, false)
	defer cancelFunc()

	t.Run("Notify", func(t *testing.T) {
		c, cancel := notify.ConnectTestClient(t)
		defer cancel()
		nn := &notifier{ApiClient: c}
		n, err := nn.createNotify()
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
		n([]warpdrive.Notification{
			{
				Peer: warpdrive.Peer{
					Id:    "Peer",
					Alias: "Alias",
					Mod:   "Mod",
				},
				Offer: warpdrive.Offer{
					OfferStatus: warpdrive.OfferStatus{
						Id:       "Id",
						In:       true,
						Status:   warpdrive.StatusUpdated,
						Index:    1,
						Progress: 1,
						Update:   1,
					},
					Create: 1,
					Peer:   "Peer",
					Files:  nil,
				},
				Info: &warpdrive.Info{
					Uri:   "",
					Path:  "",
					Size:  0,
					IsDir: false,
					Perm:  0,
					Mime:  "",
					Name:  "",
				},
			},
		})
		time.Sleep(time.Millisecond)
	})
}
