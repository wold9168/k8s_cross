// Package headscale provides a client for interacting with the Headscale API.
// This client is designed to support cross-cluster connectivity based on KEP-1645.
package headscale

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HeadscaleClient interface defines the methods that need to be implemented for interacting with Headscale.
type HeadscaleClient interface {
	GetNode(ctx context.Context, nodeId string) (*Node, error)
	ListNodes(ctx context.Context, userFilter string) ([]Node, error)
	Health(ctx context.Context) (*HealthResponse, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
}

// Client represents a client for the Headscale API.
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new Headscale API client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Node represents a node in the Headscale network.
type Node struct {
	ID           string    `json:"id"`
	MachineKey   string    `json:"machineKey"`
	NodeKey      string    `json:"nodeKey"`
	DiscoKey     string    `json:"discoKey"`
	IPAddresses  []string  `json:"ipAddresses"`
	Name         string    `json:"name"`
	User         User      `json:"user"`
	LastSeen     time.Time `json:"lastSeen"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"createdAt"`
	RegisterMethod string  `json:"registerMethod"`
	Online       bool      `json:"online"`
	ApprovedRoutes []string `json:"approvedRoutes"`
	AvailableRoutes []string `json:"availableRoutes"`
}

// User represents a user in the Headscale system.
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// ListNodesResponse represents the response from the ListNodes API endpoint.
type ListNodesResponse struct {
	Nodes []Node `json:"nodes"`
}

// GetNodeResponse represents the response from the GetNode API endpoint.
type GetNodeResponse struct {
	Node Node `json:"node"`
}

// GetNode retrieves a specific node by ID from Headscale.
func (c *Client) GetNode(ctx context.Context, nodeId string) (*Node, error) {
	url := fmt.Sprintf("%s/api/v1/node/%s", c.BaseURL, nodeId)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var getResp GetNodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&getResp); err != nil {
		return nil, err
	}
	
	return &getResp.Node, nil
}

// ListNodes retrieves all nodes from Headscale.
func (c *Client) ListNodes(ctx context.Context, userFilter string) ([]Node, error) {
	url := fmt.Sprintf("%s/api/v1/node", c.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if userFilter != "" {
		q := req.URL.Query()
		q.Add("user", userFilter)
		req.URL.RawQuery = q.Encode()
	}
	
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var listResp ListNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, err
	}
	
	return listResp.Nodes, nil
}

// HealthResponse represents the response from the health API endpoint.
type HealthResponse struct {
	DatabaseConnectivity bool `json:"databaseConnectivity"`
}

// Health checks the health status of the Headscale server.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	url := fmt.Sprintf("%s/api/v1/health", c.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		return nil, err
	}
	
	return &healthResp, nil
}

// CreateUserRequest represents the request for creating a new user.
type CreateUserRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

// CreateUserResponse represents the response from the CreateUser API endpoint.
type CreateUserResponse struct {
	User User `json:"user"`
}

// CreateUser creates a new user in Headscale.
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	url := fmt.Sprintf("%s/api/v1/user", c.BaseURL)
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	
	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create user failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	var createUserResp CreateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&createUserResp); err != nil {
		return nil, err
	}
	
	return &createUserResp.User, nil
}