package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/RobustPerception/azure_metrics_exporter/config"
	"github.com/patrickmn/go-cache"
	"strings"
)

// AzureClient represents our client to talk to the Azure api
type AzureClient struct {
	authorizer autorest.Authorizer
}

// NewAzureClient returns an Azure client to talk the Azure API
func NewAzureClient() *AzureClient {
	return &AzureClient{}
}

func (ac *AzureClient) getAuthorizer() error {
	var err error
	ac.authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return fmt.Errorf("Error getting authorizer: %v", err)
	}
	return nil
}

func (ac *AzureClient) getMetricDefinitions() (map[string]insights.MetricDefinitionCollection, error) {
	definitions := make(map[string]insights.MetricDefinitionCollection)

	// TODO: Grab the Subscription ID from wherever the Authorizer does. OR From Config File.
	client := insights.NewMetricDefinitionsClient(sc.C.Credentials.SubscriptionID)
	client.Authorizer = ac.authorizer
	client.AddToUserAgent("azure_prometheus_exporter")

	for _, target := range sc.C.Targets {
		metricsResource := fmt.Sprintf("/subscriptions/%s%s", sc.C.Credentials.SubscriptionID, target.Resource)

		var def insights.MetricDefinitionCollection
		var err error
		def, err = client.List(context.Background(), metricsResource, "")
		if err != nil {
			return nil, fmt.Errorf("Error retrieving metric definitions: %v", err)
		}

		definitions[target.Resource] = def
	}
	return definitions, nil
}
func (ac *AzureClient) getResources(searchFilter string) ([]string, error) {
	cacheKey := "resources_" + searchFilter

	cachedValue, found := memcache.Get(cacheKey)
	if found {
		return *cachedValue.(*[]string), nil
	} else {

		client := resources.NewClient(sc.C.Credentials.SubscriptionID)
		client.Authorizer = ac.authorizer
		client.AddToUserAgent("azure_prometheus_exporter")

		result, err := client.ListComplete(context.Background(), searchFilter, "resourceTypes/ID", nil)
		// TODO: Add Debug Logging
		//fmt.Println("Result: %v", result)
		//fmt.Println("Error: %v", err)
		if err != nil {
			return make([]string, 0), fmt.Errorf("Error retrieving resources: %v", err)
		}

		var output = make([]string, 0)
		//fmt.Println("Output: %v",result)
		for result.NotDone() {
			resource := result.Value()
			//fmt.Println("Result: %v", resource)
			output = append(output, *resource.ID)
			result.Next()

		}
		memcache.Set(cacheKey, &output, cache.DefaultExpiration)
		return output, nil
	}
}
func (ac *AzureClient) getMetricValue(metricNames []string, target config.Target, resourceID string) (insights.Response, error) {
	// TODO: Grab the Subscription ID from wherever the Authorizer does. OR From Config File.
	client := insights.NewMetricsClient(sc.C.Credentials.SubscriptionID)
	client.Authorizer = ac.authorizer
	client.AddToUserAgent("azure_prometheus_exporter")
	endTime, startTime := GetTimes()
	timespan := fmt.Sprintf("%s/%s", startTime, endTime)
	var aggregations string
	if len(target.Aggregations) > 0 {
		aggregations = strings.Join(target.Aggregations, ",")
	} else {
		aggregations = "Total,Average,Minimum,Maximum"
	}
	result, err := client.List(context.Background(), resourceID, timespan, nil, strings.Join(metricNames, ","), aggregations, nil, "", "", insights.Data, "")
	if err != nil {
		return insights.Response{}, fmt.Errorf("Error retrieving metrics: %v", err)
	}

	return result, nil
}
