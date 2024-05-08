package auction

import (
		"bytes"
    "context"
		"encoding/json"
    "errors"
		"net/http"
    "net/http/httptest"
		"sync"
    "testing"
    "time"

    "github.com/chrisdamba/bidders-and-auctioneers/pkg/types"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockBiddingServiceClient struct {
  mock.Mock
}

// MockAuctionService for testing without implementing business logic.
type MockAuctionService struct{}

func (m *MockBiddingServiceClient) CallBiddingService(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	args := m.Called(ctx, adRequest)
	return args.Get(0).(*models.AdObject), args.Error(1)
}

func (m *MockAuctionService) RunAuction(ctx context.Context, adRequest models.AdRequest) (*models.AdObject, error) {
	return &models.AdObject{AdID: "test-ad", BidPrice: 100.0}, nil
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
	// arrange
	mockService1 := new(MockBiddingServiceClient)
	mockService2 := new(MockBiddingServiceClient)
	adRequest := models.AdRequest{AdPlacementId: "placement1"}

	mockService1.On("CallBiddingService", mock.Anything, adRequest).Return((*models.AdObject)(nil), errors.New("no bid"))
	mockService2.On("CallBiddingService", mock.Anything, adRequest).Return((*models.AdObject)(nil), errors.New("no bid"))

	service := NewSimpleAuctionService([]BiddingServiceClient{mockService1, mockService2}, 200*time.Millisecond)
	// act
	result, err := service.RunAuction(context.Background(), adRequest)

	// assert
	assert.NotNil(t, err)
	assert.Nil(t, result)
	mockService1.AssertExpectations(t)
	mockService2.AssertExpectations(t)
}

func TestConcurrentBiddingCalls(t *testing.T) {
	// arrange
	mockService1 := &MockBiddingServiceClient{}
	mockService2 := &MockBiddingServiceClient{}


	mockService1.On("CallBiddingService", mock.Anything, mock.Anything).Return(
			&models.AdObject{BidPrice: 10}, nil).After(50 * time.Millisecond)
	mockService2.On("CallBiddingService", mock.Anything, mock.Anything).Return(
			&models.AdObject{BidPrice: 5}, nil).After(20 * time.Millisecond)


	auctionService := NewSimpleAuctionService([]BiddingServiceClient{mockService1, mockService2}, 100 * time.Millisecond) 

	adRequest := models.AdRequest{AdPlacementId: "test-placement"}

	// act
	var wg sync.WaitGroup 
	wg.Add(1)
	go func() {
			defer wg.Done()
			_, _ = auctionService.RunAuction(context.Background(), adRequest) 
	}()

	wg.Wait()

	// assert
	mockService1.AssertCalled(t, "CallBiddingService", mock.Anything, adRequest)
	mockService2.AssertCalled(t, "CallBiddingService", mock.Anything, adRequest)
}

func TestValidBidsHandling(t *testing.T) {
	// arrange
	mockService1 := &MockBiddingServiceClient{}
	mockService2 := &MockBiddingServiceClient{} 
	mockService1.On("CallBiddingService", mock.Anything, mock.Anything).Return(
			&models.AdObject{AdID: "ad1", BidPrice: 12.50}, nil) 
	mockService2.On("CallBiddingService", mock.Anything, mock.Anything).Return(
			&models.AdObject{AdID: "ad2", BidPrice: 8.00}, nil) 
	auctionService := NewSimpleAuctionService([]BiddingServiceClient{mockService1, mockService2}, 100 * time.Millisecond) 
	adRequest := models.AdRequest{AdPlacementId: "test-placement"}

	// act 
	result, err := auctionService.RunAuction(context.Background(), adRequest)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, "ad1", result.AdID)       // Expect the higher bid 
	assert.Equal(t, 12.50, result.BidPrice) 
}

func TestAdPlacementIdHandling(t *testing.T) {
	// arrange
	service := &MockAuctionService{}
	handler := NewHTTPHandler(service)

	testCases := []struct {
		description string
		adPlacementId string
		expectedAdID string
		expectedStatusCode int
	}{
		{"Valid AdPlacementId", "12345", "test-ad", http.StatusOK},
		{"Empty AdPlacementId", "", "", http.StatusBadRequest},
	}
	// act
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			requestBody, err := json.Marshal(models.AdRequest{AdPlacementId: tc.adPlacementId})
			if err != nil {
					t.Fatalf("Failed to marshal json: %v", err)
			}

			request := httptest.NewRequest("POST", "/auction", bytes.NewBuffer(requestBody))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			if response.Code == http.StatusOK {
				var adObj models.AdObject
				json.NewDecoder(response.Body).Decode(&adObj)
				assert.Equal(t, tc.expectedAdID, adObj.AdID)
			}
		})
	}
}