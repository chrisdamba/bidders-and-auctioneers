package models

type AdRequest struct {
  AdPlacementId string `json:"adPlacementId"` // Unique identifier for the Ad Placement
}

type AdObject struct {
	AdID     string  `json:"adId"`
	BidPrice float64 `json:"bidPrice,omitempty"`
}
