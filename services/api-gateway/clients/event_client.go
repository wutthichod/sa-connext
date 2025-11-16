package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wutthichod/sa-connext/shared/contracts"
)

type RouteDefinition struct {
	Method string
	Path   string
}

var routes = map[string]RouteDefinition{
	"createEvent":       {Method: "POST", Path: "/events/"},
	"getEvent":          {Method: "GET", Path: "/events/%s"},
	"getAllEvents":      {Method: "GET", Path: "/events/"},
	"joinEvent":         {Method: "POST", Path: "/events/join"},
	"getEventsByUserID": {Method: "GET", Path: "/events/users/%s"},
	"deleteEvent":       {Method: "DELETE", Path: "/events/%s"},
}

type EventServiceClient struct {
	client *http.Client
	addr   string
}

func NewEventServiceClient(addr string) *EventServiceClient {
	// Add http:// prefix if not present
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

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

	// Read body first so we can use it for error messages if needed
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (status %d): %w", resp.StatusCode, err)
	}

	var createEventResp contracts.Resp
	if err := json.Unmarshal(bodyBytes, &createEventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body (status %d): %w, body: %s", resp.StatusCode, err, string(bodyBytes))
	}

	// Ensure status code is set from HTTP response
	createEventResp.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		createEventResp.Success = true
	} else {
		createEventResp.Success = false
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

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer httpRes.Body.Close()

	var res contracts.Resp
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &res, nil
}

func (c *EventServiceClient) GetAllEvents(ctx context.Context) (*contracts.Resp, error) {
	route, ok := routes["getAllEvents"]
	if !ok {
		return nil, fmt.Errorf("route 'getAllEvents' not defined")
	}

	url := c.addr + route.Path

	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer httpRes.Body.Close()

	var res contracts.Resp
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &res, nil
}

func (c *EventServiceClient) JoinEvent(ctx context.Context, req *contracts.JoinEventRequest) (*contracts.Resp, error) {
	route, ok := routes["joinEvent"]
	if !ok {
		return nil, fmt.Errorf("route 'joinEvent' not defined")
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

	// Read body first so we can use it for error messages if needed
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body (status %d): %w", resp.StatusCode, err)
	}

	var joinEventResp contracts.Resp
	if err := json.Unmarshal(bodyBytes, &joinEventResp); err != nil {
		return nil, fmt.Errorf("failed to decode response body (status %d): %w, body: %s", resp.StatusCode, err, string(bodyBytes))
	}

	// Ensure status code is set from HTTP response
	joinEventResp.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		joinEventResp.Success = true
	} else {
		joinEventResp.Success = false
	}

	return &joinEventResp, nil
}

func (c *EventServiceClient) GetEventsByUserID(ctx context.Context, userID string) (*contracts.Resp, error) {
	route, ok := routes["getEventsByUserID"]
	if !ok {
		return nil, fmt.Errorf("route 'getEventsByUserID' not defined")
	}
	url := fmt.Sprintf(c.addr+route.Path, userID)
	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer httpRes.Body.Close()

	var res contracts.Resp
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return &res, nil
}

func (c *EventServiceClient) DeleteEvent(ctx context.Context, eventID string) (*contracts.Resp, error) {
	route, ok := routes["deleteEvent"]
	if !ok {
		return nil, fmt.Errorf("route 'deleteEvent' not defined")
	}
	url := fmt.Sprintf(c.addr+route.Path, eventID)
	httpReq, err := http.NewRequestWithContext(ctx, route.Method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpRes, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpRes.Body.Close()
	var res contracts.Resp
	if err := json.NewDecoder(httpRes.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}
	return &res, nil
}
