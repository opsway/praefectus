package main

import (
	"log"

	"github.com/boodmo/praefectus/internal/rpc"
	"github.com/boodmo/praefectus/internal/server"
	"github.com/boodmo/praefectus/internal/signals"
	"github.com/boodmo/praefectus/internal/storage"
	"github.com/boodmo/praefectus/internal/workers"
)

func main() {
	isStopping := make(chan struct{})
	ps := storage.NewProcStorage()

	rpcHandler := rpc.NewRPCHandler(ps)
	if err := rpc.Register(rpcHandler); err != nil {
		log.Fatal(err)
	}

	signals.CatchSigterm(isStopping)

	apiServer := server.New(ps)
	go apiServer.Start()

	p := workers.NewPool(2, isStopping, ps)
	p.Run()
}
