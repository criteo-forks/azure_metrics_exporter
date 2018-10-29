package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	//"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/preview/monitor/mgmt/2018-03-01/insights"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/RobustPerception/azure_metrics_exporter/config"
)

// AzureMetricDefinitionResponse represents metric definition response for a given resource from Azure.
type AzureMetricDefinitionResponse struct {
	MetricDefinitionResponses []metricDefinitionResponse `json:"value"`
}
type metricDefinitionResponse struct {
	Dimensions []struct {
		LocalizedValue string `json:"localizedValue"`
		Value          string `json:"value"`
	} `json:"dimensions"`
	ID                   string `json:"id"`
	IsDimensionRequired  bool   `json:"isDimensionRequired"`
	MetricAvailabilities []struct {
		Retention string `json:"retention"`
		TimeGrain string `json:"timeGrain"`
	} `json:"metricAvailabilities"`
	Name struct {
		LocalizedValue string `json:"localizedValue"`
		Value          string `json:"value"`
	} `json:"name"`
	PrimaryAggregationType string `json:"primaryAggregationType"`
	ResourceID             string `json:"resourceId"`
	Unit                   string `json:"unit"`
}

// AzureMetricValueResponse represents a metric value response for a given metric definition.
type AzureMetricValueResponse struct {
	Value []struct {
		Timeseries []struct {
			Data []struct {
				TimeStamp string  `json:"timeStamp"`
				Total     float64 `json:"total"`
				Average   float64 `json:"average"`
				Minimum   float64 `json:"minimum"`
				Maximum   float64 `json:"maximum"`
			} `json:"data"`
		} `json:"timeseries"`
		ID   string `json:"id"`
		Name struct {
			LocalizedValue string `json:"localizedValue"`
			Value          string `json:"value"`
		} `json:"name"`
		Type string `json:"type"`
		Unit string `json:"unit"`
	} `json:"value"`
	APIError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// AzureClient represents our client to talk to the Azure api
type AzureClient struct {
	client               *http.Client
	authorizer           autorest.Authorizer
	accessToken          string
	accessTokenExpiresOn time.Time
}

// NewAzureClient returns an Azure client to talk the Azure API
func NewAzureClient() *AzureClient {
	return &AzureClient{
		client:               &http.Client{},
		accessToken:          "",
		accessTokenExpiresOn: time.Time{},
	}
}

func (ac *AzureClient) getAuthorizer() error {
	var err error
	ac.authorizer, err = auth.NewAuthorizerFromEnvironment()
        fmt.Printf("%v",ac.authorizer)
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

func (ac *AzureClient) getMetricValue(metricNames []string, target config.Target) (insights.Response, error) {
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
        resourceUri := fmt.Sprintf("/subscriptions/%s%s",sc.C.Credentials.SubscriptionID,target.Resource)
        fmt.Printf(resourceUri)
	result, err := client.List(context.Background(), resourceUri, timespan, nil, strings.Join(metricNames, ","), aggregations, nil, "", "", insights.Data, "")
	if err != nil {
		return insights.Response{}, fmt.Errorf("Error retrieving metrics: %v", err)
	}

	return result, nil
}
