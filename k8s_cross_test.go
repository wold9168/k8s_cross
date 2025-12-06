package k8s_cross

import (
	"context"
	"fmt"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/miekg/dns"

	"github.com/wold9168/k8s_cross/headscale"
)

// TestK8sCross tests the k8s_cross plugin's basic functionality
func TestK8sCross(t *testing.T) {
	// Create a mock client using our mock implementation
	mockClient := &MockHeadscaleClient{
		Nodes: []headscale.Node{
			{
				ID:          "1",
				Name:        "my-service-1",
				IPAddresses: []string{"10.0.0.1"},
			},
		},
	}

	// Create a new K8sCross plugin instance with mock client
	x := K8sCross{
		Next:            test.ErrorHandler(),
		HeadscaleClient: mockClient,
		Zones:           []string{"clusterset.local"},
		TTL:             300,
		ClusterName:     "test-cluster",
		ClusterSet:      "test-clusterset",
	}

	ctx := context.Background()
	r := new(dns.Msg)
	r.SetQuestion("my-service.my-namespace.svc.clusterset.local.", dns.TypeA)

	// Create a Recorder to simulate DNS response writer
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// Call the plugin's ServeDNS method to handle the request
	_, err := x.ServeDNS(ctx, rec, r)
	if err != nil {
		t.Errorf("Error handling DNS request: %v", err)
	}

	// Verify the plugin handled the clusterset.local domain correctly
	if rec.Msg == nil {
		t.Error("Expected response message, got nil")
	} else if rec.Msg.Rcode != dns.RcodeSuccess {
		t.Errorf("Expected RcodeSuccess, got %d", rec.Msg.Rcode)
	}
}

// TestK8sCross_ParseClusterSetDomain tests the domain parsing functionality
func TestK8sCross_ParseClusterSetDomain(t *testing.T) {
	x := K8sCross{}

	// Test valid clusterset domain
	service, namespace, valid := x.parseClusterSetDomain("my-service.my-namespace.svc.clusterset.local.")
	if !valid {
		t.Error("Expected valid domain, got invalid")
	}
	if service != "my-service" {
		t.Errorf("Expected service 'my-service', got '%s'", service)
	}
	if namespace != "my-namespace" {
		t.Errorf("Expected namespace 'my-namespace', got '%s'", namespace)
	}

	// Test invalid domain
	_, _, valid = x.parseClusterSetDomain("invalid.domain.com.")
	if valid {
		t.Error("Expected invalid domain, got valid")
	}

	// Test domain with trailing dot
	service, namespace, valid = x.parseClusterSetDomain("service.namespace.svc.clusterset.local.")
	if !valid {
		t.Error("Expected valid domain, got invalid")
	}
	if service != "service" {
		t.Errorf("Expected service 'service', got '%s'", service)
	}
	if namespace != "namespace" {
		t.Errorf("Expected namespace 'namespace', got '%s'", namespace)
	}
}

// TestK8sCross_IsClustersetQuery tests the domain matching functionality
func TestK8sCross_IsClustersetQuery(t *testing.T) {
	x := K8sCross{Zones: []string{"clusterset.local"}}

	// Test matching domain
	if !x.isClustersetQuery("my-service.my-namespace.svc.clusterset.local.") {
		t.Error("Expected clusterset query to match")
	}

	// Test non-matching domain
	if x.isClustersetQuery("google.com.") {
		t.Error("Expected non-clusterset query to not match")
	}

	// Test with different zone
	y := K8sCross{Zones: []string{"other.local"}}
	if !y.isClustersetQuery("service.namespace.svc.other.local.") {
		t.Error("Expected other.local query to match")
	}
}

// MockHeadscaleClient is a mock implementation of the Headscale client for testing
type MockHeadscaleClient struct {
	Nodes []headscale.Node
}

func (m *MockHeadscaleClient) GetNode(ctx context.Context, nodeId string) (*headscale.Node, error) {
	// Mock implementation
	for _, node := range m.Nodes {
		if node.ID == nodeId {
			nodeCopy := node
			return &nodeCopy, nil
		}
	}
	return nil, fmt.Errorf("node not found")
}

func (m *MockHeadscaleClient) ListNodes(ctx context.Context, userFilter string) ([]headscale.Node, error) {
	return m.Nodes, nil
}

func (m *MockHeadscaleClient) Health(ctx context.Context) (*headscale.HealthResponse, error) {
	return &headscale.HealthResponse{DatabaseConnectivity: true}, nil
}

func (m *MockHeadscaleClient) CreateUser(ctx context.Context, req *headscale.CreateUserRequest) (*headscale.User, error) {
	return nil, nil
}

// TestK8sCross_FindServiceNodes tests the service node discovery functionality
func TestK8sCross_FindServiceNodes(t *testing.T) {
	// Create mock nodes
	mockNodes := []headscale.Node{
		{
			ID:          "1",
			Name:        "my-service-node1",
			IPAddresses: []string{"10.0.0.1", "2001:db8::1"},
		},
		{
			ID:          "2",
			Name:        "other-service-node2",
			IPAddresses: []string{"10.0.0.2", "2001:db8::2"},
		},
	}

	mockClient := &MockHeadscaleClient{
		Nodes: mockNodes,
	}

	x := K8sCross{
		HeadscaleClient: mockClient,
	}

	// Find nodes matching "my-service"
	nodes, err := x.findServiceNodes(context.Background(), "my-service", "")
	if err != nil {
		t.Errorf("Error finding service nodes: %v", err)
	}

	if len(nodes) == 0 {
		t.Error("Expected to find matching nodes, got none")
	} else if len(nodes) != 1 {
		t.Errorf("Expected 1 matching node, got %d", len(nodes))
	} else if nodes[0].Name != "my-service-node1" {
		t.Errorf("Expected 'my-service-node1', got '%s'", nodes[0].Name)
	}
}

// TestK8sCross_BuildRecords tests the DNS record building functionality
func TestK8sCross_BuildRecords(t *testing.T) {
	nodes := []*headscale.Node{
		{
			ID:          "1",
			Name:        "my-service-1",
			IPAddresses: []string{"10.0.0.1", "2001:db8::1"},
		},
	}

	x := K8sCross{TTL: 300}

	// Test A record building
	aRecords := x.buildARecords(nodes, "test-service", "test-namespace")
	if len(aRecords) != 1 {
		t.Errorf("Expected 1 A record, got %d", len(aRecords))
	} else if aRecords[0].Header().Rrtype != dns.TypeA {
		t.Errorf("Expected A record type, got %d", aRecords[0].Header().Rrtype)
	}

	// Test AAAA record building
	aaaaRecords := x.buildAAAARecords(nodes, "test-service", "test-namespace")
	if len(aaaaRecords) != 1 {
		t.Errorf("Expected 1 AAAA record, got %d", len(aaaaRecords))
	} else if aaaaRecords[0].Header().Rrtype != dns.TypeAAAA {
		t.Errorf("Expected AAAA record type, got %d", aaaaRecords[0].Header().Rrtype)
	}
}
