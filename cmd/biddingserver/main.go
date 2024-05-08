package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/chrisdamba/bidders-and-auctioneers/bidding"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func main() {
    var httpAddr = flag.String("http.addr", ":8080", "HTTP listen address") 
    flag.Parse()

    logger := log.NewLogfmtLogger(os.Stderr)
    logger = level.NewFilter(logger, level.AllowInfo())

    svc := bidding.NewBiddingService(logger) 
    bidEndpoint := bidding.MakeBidEndpoint(svc)
    handler := bidding.NewHTTPHandler(bidEndpoint, logger)

    // Start HTTP Server
    errs := make(chan error)
    go func() {
        c := make(chan os.Signal, 1) 
        signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
        errs <- fmt.Errorf("%s", <-c)
    }()

    go func() {
        level.Info(logger).Log("msg", "HTTP Server started", "transport", "HTTP", "addr", *httpAddr)
        errs <- http.ListenAndServe(*httpAddr, handler)
    }()

    level.Error(logger).Log("msg", "server exited", "err", <-errs)
}
