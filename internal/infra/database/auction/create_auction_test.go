package auction_test

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionCloseAutomatically(t *testing.T) {
	// Define environment variable for test duration
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Setup MongoDB connection
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI("mongodb://admin:admin@localhost:27017/auctions?authSource=admin")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Skipf("Skipping test because MongoDB is not reachable: %v", err)
		return
	}
	defer client.Disconnect(ctx)

	// Initialize Repository
	db := client.Database("auctions")
	repo := auction.NewAuctionRepository(db)

	// Create a test auction
	auctionID := "test-auction-auto-close"
	// Ensure cleanup
	db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": auctionID})
	defer db.Collection("auctions").DeleteOne(ctx, bson.M{"_id": auctionID})

	auctionEntity := &auction_entity.Auction{
		Id:          auctionID,
		ProductName: "Test Auto Close",
		Category:    "Test",
		Description: "Testing automatic closing",
		Condition:   0,
		Status:      0, // Active
		Timestamp:   time.Now(),
	}

	// Execute CreateAuction
	errErr := repo.CreateAuction(ctx, auctionEntity)
	if errErr != nil {
		t.Fatalf("Error creating auction: %v", errErr)
	}

	// Wait for the interval (2s) + buffer
	time.Sleep(3 * time.Second)

	// Verify the status in the database
	var result auction.AuctionEntityMongo
	filter := bson.M{"_id": auctionID}
	err = db.Collection("auctions").FindOne(ctx, filter).Decode(&result)
	if err != nil {
		t.Fatalf("Error retrieving auction: %v", err)
	}

	if result.Status != auction_entity.Completed {
		t.Errorf("Expected status to be Completed, got %v", result.Status)
	}
}
