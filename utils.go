package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RobustPerception/azure_metrics_exporter/config"
)

func GetSubscriptionID() string {
	if len(sc.C.Credentials.SubscriptionID) > 0 {
		fmt.Println("Using Subscription", sc.C.Credentials.SubscriptionID)
		return sc.C.Credentials.SubscriptionID
	} else {
		fmt.Println("Using Subscription", os.Getenv("AZURE_SUBSCRIPTION_ID"))
		return os.Getenv("AZURE_SUBSCRIPTION_ID")
	}
}

// PrintPrettyJSON - Prints structs nicely for debugging.
func PrintPrettyJSON(input map[string]interface{}) {
	out, err := json.MarshalIndent(input, "", "\t")
	if err != nil {
		log.Fatalf("Error indenting JSON: %v", err)
	}
	fmt.Println(string(out))
}

// GetTimes - Returns the endTime and startTime used for querying Azure Metrics API
func GetTimes() (string, string) {
	// Make sure we are using UTC
	now := time.Now().UTC()

	// Use query delay of 3 minutes when querying for latest metric data
	endTime := now.Add(time.Minute * time.Duration(-3)).Format(time.RFC3339)
	startTime := now.Add(time.Minute * time.Duration(-4)).Format(time.RFC3339)
	return endTime, startTime
}

// CreateResourceLabels - Returns resource labels for a give resource ID.
func CreateResourceLabels(resourceID string) map[string]string {
	labels := make(map[string]string)
	tmp := strings.Split(resourceID, "/")
	labels["subscription_id"] = tmp[2]
	if subscription.DisplayName != nil {
		labels["subscription_name"] = *subscription.DisplayName
	}
	labels["resource_group"] = tmp[4]
	labels["provider"] = tmp[6]
	labels["name"] = tmp[8]
	return labels
}

func hasAggregation(t config.Target, aggregation string) bool {
	// Serve all aggregations when none is specified in the config
	if len(t.Aggregations) == 0 {
		return true
	}

	for _, aggr := range t.Aggregations {
		if aggr == aggregation {
			return true
		}
	}
	return false
}
