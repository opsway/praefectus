package workers

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"sync"
)

func (p *ScalePool) listenIpcCall() {
	sockAddr := filepath.Join(os.TempDir(), fmt.Sprintf(p.config.IpcSocketPath, p.id))
	log.WithField("ScalePool::ipc", p.id).Debug("listen:", sockAddr)
	if err := os.RemoveAll(sockAddr); err != nil {
		log.WithField("ScalePool::ipc", p.id).Fatal(err)
	}
	wg := &sync.WaitGroup{}

	rpcHandler := newLiveness(p)

	rpcServer := rpc.NewServer()
	rpcServer.Register(rpcHandler)

	listener, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.WithField("ScalePool::ipc", p.id).Fatal("listen error:", err)
	}
	wg.Add(1)
	go func(wGroup *sync.WaitGroup) {
		rpcServer.Accept(listener)
		listener.Close()
		wGroup.Done()
	}(wg)
	wg.Wait()
}
