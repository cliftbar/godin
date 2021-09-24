package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"
)

// FirestoreEvent is the payload of a Firestore event.
// Please refer to the docs for additional information
// regarding Firestore events.
type FirestoreEvent struct {
	Value    FirestoreValue `json:"value"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	Name string
	StormID string
	AdvNumber int
	Timestamp time.Time
	LatY float64
	LonX float64
	BearingDeg float64
	ForwardSpeedKts float64
	VMaxKts float64
	MinCpMb float64
	Discussion string
	Graphics []string
	Sources []string
}


//HelloFirestore is triggered by a change to a Firestore document.
func HelloFirestore(ctx context.Context, e FirestoreEvent) error {
	toPubSub(e.Value)
	return nil
}

func toPubSub (data FirestoreValue) {
	projectID, ok := os.LookupEnv("GCP_PROJECT")
	if !ok {
		projectID = "godin-324403"
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
	t := client.Topic("odin-build")
	dataBytes, _ := json.Marshal(data)
	res := t.Publish(ctx, &pubsub.Message{
		Data:            dataBytes,
	})

	id, err := res.Get(ctx)
	if err != nil {
		log.Fatalf("error! %v", err)
	} else {
		log.Printf("pub sub submitted, id: %v", id)
	}
}

func main(){
	toPubSub(FirestoreValue{})
}