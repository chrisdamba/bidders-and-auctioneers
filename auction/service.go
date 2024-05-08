package auction

import (
	"context"
	"sync"
	"time"

	"github.com/chrisdamba/bidders-and-auctioneers/bidding"
	"github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
)

type BiddingServiceClient interface {
  CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error)
}

type AuctionService interface {
  RunAuction(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error)
}

type SimpleAuctionService struct {
    biddingServices []BiddingServiceClient
    auctionTimeout  time.Duration
}

func NewSimpleAuctionService(biddingServices []BiddingServiceClient, timeout time.Duration) *SimpleAuctionService {
	return &SimpleAuctionService{
		biddingServices: biddingServices,
		auctionTimeout:  timeout,
	}
}

func (s *SimpleAuctionService) RunAuction(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	ctx, cancel := context.WithTimeout(ctx, s.auctionTimeout)
	defer cancel()

	var mu sync.Mutex
	var winningBid *models.AdObject

	var wg sync.WaitGroup
	// Start bidding in parallel
	for _, service := range s.biddingServices {
		wg.Add(1)
		go func(service BiddingServiceClient) {
			defer wg.Done()
			bid, err := service.CallBiddingService(ctx, adRequest)
			if err != nil {
				return
			}
			mu.Lock()
			if winningBid == nil || bid.BidPrice > winningBid.BidPrice {
				winningBid = bid
			}
			mu.Unlock()
		}(service)
	}

	wg.Wait()

	if winningBid == nil {
		return nil, bidding.ErrNoBid
	}

	return winningBid, nil
}
