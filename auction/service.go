package auction

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/chrisdamba/bidders-and-auctioneers/bidding"
	"github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
)

type BiddingServiceClient interface {
	CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error)
}

type biddingServiceClient struct {
	baseURL string
	client  *http.Client
}

// CallBiddingService implements BiddingServiceClient.
func (b *biddingServiceClient) CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	panic("unimplemented")
}

func NewBiddingServiceClient(baseURL string) BiddingServiceClient {
	return &biddingServiceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second}, // Adjust timeout as needed
	}
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

// NewHTTPHandler returns an http.Handler for the auction service.
func NewHTTPHandler(service AuctionService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auction", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var adRequest models.AdRequest
		err := json.NewDecoder(r.Body).Decode(&adRequest)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if adRequest.AdPlacementId == "" {
			http.Error(w, "Bad Request: Missing AdPlacementId", http.StatusBadRequest)
			return
		}

		adObject, err := service.RunAuction(r.Context(), adRequest)
		if err != nil {
			if err == bidding.ErrNoBid {
				http.Error(w, "No Content: "+err.Error(), http.StatusNoContent)
				return
			}
			http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(adObject)
	})

	return mux
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
