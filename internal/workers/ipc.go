package workers

import (
	"log"
	"net"
	"net/rpc"
	"os"

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
	defer listener.Close()

	conn := make(chan net.Conn, 1)
	go func(conn chan net.Conn) {
		for {
			c, err := listener.AcceptUnix()
			if err != nil {
				log.Fatal("accept error:", err) // ToDo: Refactor?
			}
			conn <- c
		}
	}(conn)

	for {
		select {
		case <-isDone:
			if err := os.RemoveAll(path); err != nil {
				return err // ToDo: Error wrapping
			}
			return nil
		case c := <-conn:
			go rpc.ServeCodec(goridgeRpc.NewCodec(c))
		}
	}
}
