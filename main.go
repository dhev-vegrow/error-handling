package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"test-otel-app/utils/otel"
)

func processOrder(ctx context.Context, orderID string) error {
	ctx, span := otel.StartSpan(ctx, "processOrder")
	defer span.End()

	otel.LogInfo(ctx, fmt.Sprintf("Processing order %s", orderID))

	// Simulate processing steps
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	if rand.Float32() < 0.5 { // 50% chance of error
		err := fmt.Errorf("failed to process order %s", orderID)
		otel.LogError(ctx, err)
		return err
	}

	otel.LogInfo(ctx, fmt.Sprintf("Order %s processed successfully", orderID))
	return nil
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderID := fmt.Sprintf("ORDER-%d", rand.Intn(1000))

	err := processOrder(ctx, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Order %s processed successfully", orderID)
}

func main() {
	ctx := context.Background()
	cleanup := otel.InitProvider(ctx)
	defer cleanup()

	http.HandleFunc("/order", handleOrder)

	otel.LogInfo(context.Background(), "Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
