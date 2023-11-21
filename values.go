package warpdrive

import "errors"

const (
	Port    = "warpdrive"
	PortCli = "wd"
)

var ErrEnded = errors.New("ended")
