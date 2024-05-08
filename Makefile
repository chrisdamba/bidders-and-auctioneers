
### Makefile

```makefile
.PHONY: build run test clean

build:
	@echo "Building bidding service..."
	@cd cmd/biddingserver && go build -o bidding-server
	@echo "Building auction service..."
	@cd cmd/auctionserver && go build -o auction-server

run:
	@echo "Running bidding service..."
	@cd cmd/biddingserver && ./bidding-server &
	@echo "Running auction service..."
	@cd cmd/auctionserver && ./auction-server &

test:
	@echo "Running tests for bidding service..."
	@go test ./bidding/...
	@echo "Running tests for auction service..."
	@go test ./auction/...

clean:
	@echo "Cleaning up..."
	@rm cmd/biddingserver/bidding-server
	@rm cmd/auctionserver/auction-server

docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

