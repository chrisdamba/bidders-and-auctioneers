package bidding

import (
	"context"
	"testing"

	"github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
)

func TestBiddingServiceMakeBid(t *testing.T) {
	service := biddingService{}
	ctx := context.Background()

	// Running 100 iterations to account for randomness in the decision to bid or not.
	bidMade := false
  for i := 0; i < 100; i++ {
    adObject, err := service.MakeBid(ctx, models.AdRequest{AdPlacementId: "test-ad-placement"})
		if err == nil {
			bidMade = true
			if adObject.AdID == "" || adObject.BidPrice == 0 {
				t.Errorf("Expected valid AdID and BidPrice, got AdID: %s, BidPrice: %f", adObject.AdID, adObject.BidPrice)
			}
		} else if err != ErrNoBid {
			t.Errorf("Expected ErrNoBid or nil, got %v", err)
		}
  }

	if !bidMade {
		t.Errorf("Expected at least one successful bid out of 100 trials")
	}
}
