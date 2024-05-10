package auction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/chrisdamba/bidders-and-auctioneers/bidding"
	"github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
)

type BiddingServiceClient interface {
	CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error)
	GetBaseURL() string
}


type biddingServiceClient struct {
	baseURL string
	client  *http.Client
}

type AuctionService interface {
	RunAuction(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error)
}

type SimpleAuctionService struct {
	biddingServices 	[]BiddingServiceClient
	auctionTimeout  	time.Duration
	circuitBreakers   map[string]*CircuitBreaker
	logger          	*log.Logger
}


func (b *biddingServiceClient) CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	requestURL := fmt.Sprintf("%s/bid?adPlacementId=%s", b.baseURL, adRequest.AdPlacementId)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, nil) 
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to bidding service: %w", err)
	}
	defer resp.Body.Close() 

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNoContent { 
			return nil, bidding.ErrNoBid // Handle no-bid scenario
		}
		return nil, fmt.Errorf("bidding service error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var adObject models.AdObject
	err = json.Unmarshal(body, &adObject)
	if err != nil {
			return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &adObject, nil
}

func NewBiddingServiceClient(baseURL string) BiddingServiceClient {
	return &biddingServiceClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second}, // Adjust timeout as needed
	}
}


func (b *biddingServiceClient) GetBaseURL() string {
	return b.baseURL
}


func NewSimpleAuctionService(biddingServices []BiddingServiceClient, timeout time.Duration, failureThreshold int, cooldownPeriod time.Duration) *SimpleAuctionService {
	logger := log.New(os.Stdout, "auction-service: ", log.LstdFlags)
	circuitBreakers := make(map[string]*CircuitBreaker)
	for _, service := range biddingServices {
		circuitBreakers[service.GetBaseURL()] = NewCircuitBreaker(failureThreshold, cooldownPeriod) 
	}
	return &SimpleAuctionService{
		biddingServices: biddingServices,
		auctionTimeout:  timeout,
		circuitBreakers: circuitBreakers, 
		logger:          logger,
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
				s.logger.Println("bidding service error:", err) 
				return
			}
			mu.Lock()
			if winningBid == nil || bid.BidPrice > winningBid.BidPrice {
				winningBid = bid
			}
			mu.Unlock()
			s.logger.Printf("bid received: adPlacementId=%s, price=%f", adRequest.AdPlacementId, bid.BidPrice) 
		}(service)
	}

	wg.Wait()

	if winningBid == nil {
		return nil, bidding.ErrNoBid
	}

	return winningBid, nil
}
