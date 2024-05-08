# Bidding and Auction Services

This repository contains the implementation of two microservices: a Bidding Service and an Auction Service. Both services are designed to handle ad placement bidding and auction management respectively.

* **Bidding Service:** Handles incoming ad requests and generates bids.
* **Auction Service:** Conducts auctions across multiple bidding services and selects the highest bid.

**Features**

* **Header Bidding:** Supports header bidding to maximize ad revenue.
* **Concurrency:** Services are designed with concurrency in mind.
* **Safety Circuit:** A circuit breaker prevents latency issues caused by slow bidders.
* **MongoDB Integration:** Optional persistence of ad requests, bids, and auction results for analytics and reporting.
* **Observability:** Structured logging and the potential for metrics integration.


## Project Structure

The project is divided into several main directories:

- **`/bidding` and `/auction`**: Contain all the code specific to each service, including business logic (`service.go`), data models (`models.go`), and HTTP transport (`transport/http.go`).
- **`/pkg`**: Contains shared packages used across different services, such as the database connection logic (`db/db.go`) and shared types (`types/types.go`).
- **`/cmd`**: Contains the main applications for the Bidding and Auction services.
- **`/config`**: Handles configurations needed by the services.

## Prerequisites

To run this project, you will need:

- Go Go (version 1.18 or later)
- Docker and Docker Compose (optional, simplifies deployment)
- MongoDB (optional, for persistence)

## Setup

Clone the repository to your local machine:

```bash
git clone git@github.com:chrisdamba/bidders-and-auctioneers.git
cd bidders-and-auctioneers
```

## (Optional) Start MongoDB
If you plan to use MongoDB, ensure it's running. Instructions on how to set up locally can be found on [the MongoDB site](https://www.mongodb.com/docs/manual/installation/).

## Build and Run with Docker Compose (Optional)
  ```bash
  docker-compose up --build
  ```

## Build and Run Manually
  * **Build the Services:**
      ```bash
      cd cmd/biddingserver && go build
      cd ../auctionserver && go build
      ```  
  * **Run the Services:**
      ```bash
      ./biddingserver &
      ./auctionserver &
      ``` 

## Testing

1. **Unit Tests:**
   ```bash
   cd bidding && go test 
   cd ../auction && go test
   ```

## API Usage

* **Bidding Service:**
   * Endpoint: `/bid`
   * Method: `POST`
   * Body: JSON-formatted `AdRequest` object 
* **Auction Service:**
   * Endpoint: `/auction`
   * Method: `POST`
   * Body: JSON-formatted `AdRequest` object 
