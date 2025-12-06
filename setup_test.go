package k8s_cross

import (
	"testing"

	"github.com/coredns/caddy"
)

// TestParseConfig tests the configuration parsing functionality
func TestParseConfig(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedZones []string
	}{
		{
			name: "valid config with headscale_url",
			input: `k8s_cross {
    headscale_url http://headscale:8080 api-key-123
    cluster my-cluster
    clusterset my-clusterset
    ttl 600
}`,
			expectError:   false,
			expectedZones: []string{"."},
		},
		{
			name: "config with zones specified",
			input: `k8s_cross clusterset.local {
    headscale_url http://headscale:8080 api-key-123
}`,
			expectError:   false,
			expectedZones: []string{"clusterset.local"},
		},
		{
			name: "missing headscale_url",
			input: `k8s_cross {
    cluster my-cluster
}`,
			expectError: true,
		},
		{
			name: "invalid ttl value",
			input: `k8s_cross {
    headscale_url http://headscale:8080 api-key-123
    ttl invalid
}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.input)
			k8sCross, err := parseConfig(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if k8sCross.HeadscaleClient == nil {
				t.Errorf("Expected HeadscaleClient to be set")
			}

			if len(k8sCross.Zones) != len(tt.expectedZones) {
				t.Errorf("Expected %d zones, got %d", len(tt.expectedZones), len(k8sCross.Zones))
			} else {
				for i, expectedZone := range tt.expectedZones {
					if k8sCross.Zones[i] != expectedZone {
						t.Errorf("Expected zone %s at position %d, got %s", expectedZone, i, k8sCross.Zones[i])
					}
				}
			}
		})
	}
}