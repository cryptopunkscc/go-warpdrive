package service

import (
	"github.com/cryptopunkscc/go-warpdrive"
	"github.com/mitchellh/ioprogress"
	"io"
	"time"
)

func (srv *offerService) Copy(offer *warpdrive.Offer) warpdrive.CopyOffer {
	srv.Offer = offer
	return srv
}

func (srv *offerService) From(reader io.Reader) (err error) {
	offer := srv.Offer
	offer.Status = warpdrive.StatusUpdated
	for i := range offer.Files {
		if i < offer.Index {
			continue
		}
		offer.Index = i
		if err = srv.fileFrom(reader); err != nil {
			return
		}
		offer.Progress = 0
	}
	return
}

func (srv *offerService) fileFrom(reader io.Reader) (err error) {
	s := srv.fileStorage
	o := srv.Offer

	info := o.Files[o.Index]
	if info.IsDir {
		err := s.MkDir(info.Uri, info.Perm)
		if err != nil && !s.IsExist(err) {
			srv.Println("Cannot make dir", info.Uri, err)
			return err
		}
		srv.progress(info.Size)
		return nil
	}
	offset := o.Progress
	writer, err := s.FileWriter(info.Uri, info.Perm, offset)
	if err != nil {
		srv.Println("Cannot get writer for", info.Uri, err)
		return
	}
	defer func() {
		err := writer.Close()
		if err != nil {
			srv.Println("Cannot close info", info.Uri, err)
			return
		}
	}()
	// Copy bytes
	update := func(progress int64, size int64) error {
		srv.progress(offset + progress)
		return nil
	}
	progress := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progress, info.Size-offset)
	o.Progress = offset + l
	if err != nil {
		srv.Println("Cannot read", info.Uri, err, "expected size", info.Size, "but was", o.Progress)
		return err
	}
	if o.Progress != info.Size {
		srv.progress(info.Size)
	}
	return
}

func (srv *offerService) To(writer io.Writer) (err error) {
	o := srv.Offer
	o.Status = warpdrive.StatusUpdated
	for i := range o.Files {
		if i < o.Index {
			continue
		}
		o.Index = i
		if err = srv.fileTo(writer); err != nil {
			return
		}
		o.Progress = 0
	}
	return
}

func (srv *offerService) fileTo(writer io.Writer) (err error) {
	o := srv.Offer
	info := o.Files[o.Index]
	if info.IsDir {
		srv.progress(info.Size)
		return
	}
	offset := o.Progress
	reader, err := srv.resolver.Reader(info.Uri, offset)
	if err != nil {
		srv.Println("Cannot get reader", info.Uri, o.Id, err)
		return
	}
	defer reader.Close()
	update := func(progress int64, size int64) error {
		srv.progress(offset + progress)
		return nil
	}
	progressReader := &ioprogress.Reader{
		Reader:       reader,
		Size:         info.Size,
		DrawInterval: 1000 * time.Millisecond,
		DrawFunc:     update,
	}
	l, err := io.CopyN(writer, progressReader, info.Size-offset)
	if err != nil {
		srv.Println("Cannot write", info.Uri, err)
		return err
	}
	o.Progress = offset + l
	if o.Progress != info.Size {
		srv.progress(info.Size)
	}
	return
}

func (srv *offerService) progress(progress int64) {
	srv.Offer.Progress = progress
	srv.update(srv.Offer)
}
