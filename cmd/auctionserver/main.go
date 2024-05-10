package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/chrisdamba/bidders-and-auctioneers/auction"
)

func main() {
	var (
		httpAddr       		= flag.String("http.addr", ":8081", "HTTP listen address")
		auctionTimeout 		= flag.Duration("auction.timeout", 100*time.Millisecond, "Timeout for the auction itself")
		failureThreshold 	= flag.Int("auction.failureThreshold", 3, "Failure threshold for the auction") 
		cooldownPeriod   	= flag.Duration("auction.cooldownPeriod", 100 * time.Millisecond, "Cooldown period for the auction")
	)
	flag.Parse()

	logger := log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowInfo()) 

	biddingServiceClients := []auction.BiddingServiceClient{
		auction.NewBiddingServiceClient("http://localhost:8080"),
		auction.NewBiddingServiceClient("http://localhost:8081"),
		auction.NewBiddingServiceClient("http://localhost:8082"),
	}

	// Create auction service 
	svc := auction.NewSimpleAuctionService(biddingServiceClients, *auctionTimeout, *failureThreshold, *cooldownPeriod) 

	httpHandler := auction.NewHTTPHandler(svc)

	// Start HTTP Server
	httpServer := &http.Server{
		Handler:      httpHandler,
		Addr:         *httpAddr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal, 1) 
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		level.Info(logger).Log("msg", "HTTP Server started", "transport", "HTTP", "addr", *httpAddr)
		errs <- httpServer.ListenAndServe()
	}()

	level.Error(logger).Log("msg", "server exited", "err", <-errs)
}
