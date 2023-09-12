package dhcpsvc

import (
	"sync/atomic"
)

type DHCPServer struct {
	enabled *atomic.Bool
}

// func New(conf *Config) (srv *DHCPServer, err error) {
// }
