package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wutthichod/sa-connext/shared/contracts"
)

type RouteDefinition struct {
	Method string
	Path   string
}

var routes = map[string]RouteDefinition{
	"createEvent": RouteDefinition{Method: "POST", Path: "/"},
	"getEvent":    RouteDefinition{Method: "GET", Path: "/"},
}

type EventServiceClient struct {
	client *http.Client
	addr   string
}

func NewEventServiceClient(addr string) *EventServiceClient {
	return &EventServiceClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		addr: addr,
	}
}

func (c *EventServiceClient) CreateEvent(ctx context.Context, req *contracts.CreateEventRequest) (*contracts.Resp, error) {
	route, ok := routes["createEvent"]
	if !ok {
		return nil, fmt.Errorf("route 'createEvent' not defined")
	}

	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := c.addr + route.Path
	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	var createEventResp contracts.Resp
	if err := json.NewDecoder(resp.Body).Decode(&createEventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &createEventResp, nil
}

func (c *EventServiceClient) GetEventById(ctx context.Context, eventID string) (*contracts.Resp, error) {
	route, ok := routes["getEvent"]
	if !ok {
		return nil, fmt.Errorf("route 'getEvent' not defined")
	}

	url := fmt.Sprintf(c.addr+route.Path, eventID)

	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	var getEventResp contracts.Resp
	if err := json.NewDecoder(resp.Body).Decode(&getEventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &getEventResp, nil
}

func (c *EventServiceClient) JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (*contracts.Resp, error) {
	route, ok := routes["joinEvent"]
	if !ok {
		return nil, fmt.Errorf("route 'createEvent' not defined")
	}

	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := c.addr + route.Path
	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	var createEventResp contracts.Resp
	if err := json.NewDecoder(resp.Body).Decode(&createEventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &createEventResp, nil
}
