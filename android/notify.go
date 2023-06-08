package android

import (
	"fmt"
	"go-warpdrive/adapter"
	android "go-warpdrive/android/notify"
	"go-warpdrive/proto"
	"go-warpdrive/service"
	"log"
	"strconv"
	"strings"
)

type notifier struct {
	android.Client
	notify        android.Notify
	inChannel     android.Channel
	outChannel    android.Channel
	inGroup       android.Notification
	outGroup      android.Notification
	lastId        int
	notifications map[proto.OfferId]*android.Notification
}

func NewNotifier(api adapter.Api) service.Notify {
	m := &notifier{}
	m.Api = api
	m.inChannel = android.Channel{
		Id:         "warpdrive-in",
		Name:       "Warp Drive incoming",
		Importance: android.ImportanceMax,
	}
	m.outChannel = android.Channel{
		Id:         "warpdrive-out",
		Name:       "Warp Drive outgoing",
		Importance: android.ImportanceDefault,
	}
	groupCommon := android.Notification{
		SubText:       "Warp Drive",
		OnlyAlertOnce: true,
		AutoCancel:    true,
		Silent:        true,
		GroupSummary:  true,
	}
	m.inGroup = groupCommon
	m.inGroup.Id = m.nextId()
	m.inGroup.ChannelId = m.inChannel.Id
	m.inGroup.Group = "in"

	m.outGroup = groupCommon
	m.outGroup.Id = m.nextId()
	m.outGroup.ChannelId = m.outChannel.Id
	m.outGroup.Group = "out"

	m.notifications = map[proto.OfferId]*android.Notification{}

	m.createChannels()

	return m.Notify
}

func (m *notifier) createChannels() *notifier {

	err := m.Create(m.inChannel)
	if err != nil {
		log.Println("Cannot create incoming notification channel", err)
		return nil
	}
	err = m.Create(m.outChannel)
	if err != nil {
		log.Println("Cannot create outgoing notification channel", err)
		return nil
	}
	m.notify = m.Notifier()
	return m
}

func (m *notifier) Notify(notifications []service.Notification) {
	// Update notifications cache
	for _, n := range notifications {
		switch n.Status {
		case proto.StatusAwaiting:
			if n.In && n.Peer.Mod == proto.PeerModAsk {
				m.create(n)
			}
		case proto.StatusUpdated:
			m.progress(n)
		case
			proto.StatusFailed,
			proto.StatusRejected,
			proto.StatusCompleted:
			m.finish(n)
		}
	}

	// Collect notifications for send
	var arr []android.Notification
	for _, group := range []android.Notification{m.inGroup, m.outGroup} {
		var last *android.Notification
		var exist bool
		for _, n := range notifications {
			last, exist = m.notifications[n.Offer.Id]
			if exist && last.ChannelId == group.ChannelId {
				arr = append(arr, *last)
			}
		}
		// Add notification group if needed
		if exist && last.ChannelId == group.ChannelId {
			arr = append(arr, group)
		}
	}

	// Dispatch notifications
	if m.notify != nil {
		m.notify <- arr
	}

	// remove finished notifications
	for _, n := range notifications {
		switch n.Status {
		case
			proto.StatusFailed,
			proto.StatusRejected,
			proto.StatusCompleted:
			delete(m.notifications, n.Offer.Id)
		}
	}
}

func (m *notifier) create(an service.Notification) (n *android.Notification) {
	channel := m.outChannel
	n = &android.Notification{
		Id:            m.nextId(),
		ChannelId:     channel.Id,
		Ongoing:       false,
		OnlyAlertOnce: true,
		AutoCancel:    true,
		Priority:      android.PriorityMax,
		SubText:       "Warp Drive",
		Number:        len(an.Files),
		ContentIntent: &android.Intent{
			Uri: "warpdrive://offer/" + string(an.Offer.Id),
		},
	}
	if an.In {
		n.Group = "in"
		n.Action = &android.Action{
			Title: "download",
			Intent: &android.Intent{
				Uri: "warpdrive://download/" + string(an.Offer.Id),
			},
		}
	} else {
		n.Group = "out"
	}
	if an.In {
		n.ChannelId = m.inChannel.Id
		peerName := an.Peer.Alias
		if peerName == "" {
			shortId := string(an.Peer.Id)[58:66]
			peerName = shortId[0:4] + "-" + shortId[4:8]
		}
		filename := ""
		size := int64(0)
		if len(an.Files) > 0 {
			filename = an.Files[0].Uri
		}
		for _, info := range an.Files {
			size += info.Size
			if !strings.HasPrefix(info.Uri, filename) {
				filename = ""
			}
		}
		title := peerName + " wants to share"
		text := ""
		if len(an.Files) > 1 {
			if filename != "" {
				title += " a directory " + filename
			}
			text = strconv.Itoa(len(an.Files)) + " files with summary size "
		} else {
			title += " a file " + filename
			text = "with size "
		}
		text += ByteCountSI(size)
		n.ContentTitle = title
		n.ContentText = text
	}
	m.notifications[an.Offer.Id] = n
	return
}

func (m *notifier) progress(an service.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
	}
	if an.Info == nil {
		log.Println("Cannot update progress for nil Info")
		return
	}
	n.Ongoing = true
	n.AutoCancel = false
	n.ContentTitle = titlePrefix(an) + " " + formatPeerName(an)
	n.ContentText = an.Uri + " " + ByteCountSI(an.Progress) + " / " + ByteCountSI(an.Size)
	n.Progress = &android.Progress{
		Max:     int(an.Size),
		Current: int(an.Progress),
	}
	n.Action = nil
	m.notifications[an.Offer.Id] = n
}

func (m *notifier) finish(an service.Notification) {
	n := m.notifications[an.Offer.Id]
	if n == nil {
		n = m.create(an)
	}
	n.Ongoing = false
	n.AutoCancel = true
	n.ContentTitle = titlePrefix(an) + " " + formatPeerName(an) + " " + an.OfferStatus.Status
	n.ContentText = fmt.Sprintf(
		"transferred %d/%d files with size %s",
		an.Index,
		len(an.Files),
		formatTransferredSize(an),
	)
	n.Progress = nil
	m.notifications[an.Offer.Id] = n
}

func titlePrefix(an service.Notification) string {
	if an.In {
		return "Downloading from"
	}
	return "Uploading to"
}

func formatPeerName(an service.Notification) (name string) {
	name = an.Peer.Alias
	if name == "" && len(an.Peer.Id) == 66 {
		shortId := string(an.Peer.Id)[58:66]
		name = shortId[0:4] + "-" + shortId[4:8]
	}
	if name == "" {
		name = "this device"
	}
	return
}

func formatTransferredSize(an service.Notification) (str string) {
	str = ByteCountSI(sumSize(an))
	if an.Index < len(an.Files) {
		str += " / " + ByteCountSI(totalSize(an))
	}
	return
}

func totalSize(an service.Notification) (size int64) {
	for _, file := range an.Files {
		size += file.Size
	}
	return
}

func sumSize(an service.Notification) (size int64) {
	for i := 0; i < an.Index; i++ {
		size += an.Files[i].Size
	}
	size += an.Progress
	return
}

func (m *notifier) nextId() int {
	m.lastId++
	return m.lastId
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}