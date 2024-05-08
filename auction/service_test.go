package auction

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockBiddingServiceClient struct {
    mock.Mock
}

func (m *MockBiddingServiceClient) CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	args := m.Called(ctx, adRequest)
	return args.Get(0).(*models.AdObject), args.Error(1)
}

func TestRunAuction(t *testing.T) {
	mockService1 := new(MockBiddingServiceClient)
	mockService2 := new(MockBiddingServiceClient)
	adRequest := models.AdRequest{AdPlacementId: "placement1"}

	mockService1.On("CallBiddingService", mock.Anything, adRequest).Return(&models.AdObject{AdID: "ad1", BidPrice: 10.0}, nil)
	mockService2.On("CallBiddingService", mock.Anything, adRequest).Return(&models.AdObject{AdID: "ad2", BidPrice: 15.0}, nil)

	service := NewSimpleAuctionService([]BiddingServiceClient{mockService1, mockService2}, 200*time.Millisecond)
	result, err := service.RunAuction(context.Background(), adRequest)

	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 15.0, result.BidPrice)
	mockService1.AssertExpectations(t)
	mockService2.AssertExpectations(t)
}

func TestRunAuctionNoBids(t *testing.T) {
	mockService1 := new(MockBiddingServiceClient)
	mockService2 := new(MockBiddingServiceClient)
	adRequest := models.AdRequest{AdPlacementId: "placement1"}

	mockService1.On("CallBiddingService", mock.Anything, adRequest).Return(nil, errors.New("no bid"))
	mockService2.On("CallBiddingService", mock.Anything, adRequest).Return(nil, errors.New("no bid"))

	service := NewSimpleAuctionService([]BiddingServiceClient{mockService1, mockService2}, 200*time.Millisecond)
	result, err := service.RunAuction(context.Background(), adRequest)

	assert.NotNil(t, err)
	assert.Nil(t, result)
	mockService1.AssertExpectations(t)
	mockService2.AssertExpectations(t)
}
