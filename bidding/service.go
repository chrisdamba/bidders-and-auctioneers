package bidding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/log"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
)

type BiddingService interface {
	MakeBid(ctx context.Context, adRequest models.AdRequest) (models.AdObject, error)
}

type biddingService struct{}

func (b biddingService) MakeBid(ctx context.Context, adRequest models.AdRequest) (models.AdObject, error) {
	// Randomly decide not to bid
	if rand.Float32() < 0.5 {
		return models.AdObject{}, ErrNoBid
	}

	// Generate a random bid
	return models.AdObject{
		AdID:     generateAdID(),
		BidPrice: rand.Float64() * 10,
	}, nil
}

var ErrNoBid = errors.New("no bid made")

func generateAdID() string {
  return fmt.Sprintf("ad-%d", rand.Int())
}


func MakeBidEndpoint(svc BiddingService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(bidRequest)
		adRequest := models.AdRequest{AdPlacementId: req.AdPlacementId}
		adObject, err := svc.MakeBid(ctx, adRequest)
		if err != nil {
			return bidResponse{adObject, err.Error()}, nil
		}
		return bidResponse{adObject, ""}, nil
	}
}

type bidRequest struct {
	AdPlacementId string
}

type bidResponse struct {
  models.AdObject
  Err string `json:"err,omitempty"`
}


func decodeHTTPBidRequest(_ context.Context, r *http.Request) (interface{}, error) {
  var req bidRequest
  req.AdPlacementId = r.URL.Query().Get("adPlacementId")
  return req, nil
}

func encodeHTTPResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(endpoint.Failer); ok && e.Failed() != nil {
		encodeError(ctx, e.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == ErrNoBid {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func NewHTTPHandler(endpoint endpoint.Endpoint, logger log.Logger) http.Handler {
	m := http.NewServeMux()
	m.Handle("/bid", httptransport.NewServer(
		endpoint,
		decodeHTTPBidRequest,
		encodeHTTPResponse,
		httptransport.ServerErrorLogger(logger),
	))
	return m
}
