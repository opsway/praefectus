package workers

import (
	"net"
	"net/rpc"
	"os"

	log "github.com/sirupsen/logrus"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
)

func listenUnixSocket(path string, isDone chan struct{}) error {
	unixAddr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return err
	}

	listener, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		return err
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.WithError(err).Error("IPC close error")
		}
	}()

	conn := make(chan net.Conn, 1)
	go func(conn chan net.Conn) {
		for {
			c, err := listener.AcceptUnix()
			if err != nil {
				log.WithError(err).Error("IPC accept error")
				continue
			}
			conn <- c
		}
	}(conn)

	for {
		select {
		case <-isDone:
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			return nil
		case c := <-conn:
			go rpc.ServeCodec(goridgeRpc.NewCodec(c))
		}
	}
}
