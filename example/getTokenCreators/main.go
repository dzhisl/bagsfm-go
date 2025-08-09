package main

import (
	"context"
	"encoding/json"
	"log"

	bags "github.com/dzhisl/bagsfm-go"
)

func main() {
	apiKey := "your-api-key"
	client, err := bags.New(apiKey, nil)
	if err != nil {
		log.Fatalf("failed to create bags client: %s", err)
	}
	creators, err := client.GetTokenLaunchCreators(context.Background(), "5qSVmtYCNmsEpktudHJCoUcHPEqmY9TN2xwv59NJBAGS")
	if err != nil {
		log.Fatalf("failed to get token creators: %s", err)
	}
	for i, creator := range creators {
		jsoned, err := json.Marshal(creator)
		if err != nil {
			log.Fatalf("failed to marshal token creator: %s", err)
		}
		log.Printf("Token creator: %d\n", i)
		log.Printf("Info: %s", jsoned)
	}

}
