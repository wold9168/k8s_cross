package k8s_cross

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/wold9168/k8s_cross/headscale"
)

// init registers this plugin with CoreDNS.
func init() {
	caddy.RegisterPlugin("k8s_cross", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

// setup is the function that gets called when the plugin is loaded. It configures the plugin.
// The function receives a *caddy.Controller which provides access to the configuration.
func setup(c *caddy.Controller) error {
	// Parse the configuration from the CoreDNS config file
	k8sCross, err := parseConfig(c)
	if err != nil {
		return plugin.Error("k8s_cross", err)
	}

	// Add the plugin to the DNS server middleware chain
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		k8sCross.Next = next
		return k8sCross
	})

	return nil
}

// parseConfig parses the configuration for the k8s_cross plugin from the CoreDNS config.
func parseConfig(c *caddy.Controller) (*K8sCross, error) {
	var k8sCross *K8sCross

	// Parse the configuration block
	for c.Next() {
		// The first argument after the plugin name is the zones for which this plugin is responsible
		zones := c.RemainingArgs()
		if len(zones) == 0 {
			zones = []string{"."} // Default to all zones if none specified
		}

		// Initialize the plugin instance
		k8sCross = &K8sCross{
			Zones:       zones,
			TTL:         300, // Default TTL of 5 minutes
			ClusterName: "default-cluster",
			ClusterSet:  "default-clusterset",
		}

		// Parse configuration options
		for c.NextBlock() {
			switch c.Val() {
			case "headscale_url":
				// Parse Headscale API URL and authentication
				args := c.RemainingArgs()
				if len(args) < 2 {
					return nil, fmt.Errorf("headscale_url requires both URL and API key")
				}
				url := args[0]
				apiKey := args[1]

				// Create Headscale client
				client := headscale.NewClient(url, apiKey)
				k8sCross.HeadscaleClient = client

				// Test the connection
				// Note: In a production implementation, you might want to defer this check
				// until the plugin is actually used to avoid startup delays
			case "ttl":
				// Parse custom TTL value
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, fmt.Errorf("ttl requires exactly one argument")
				}
				var ttl uint32
				_, err := fmt.Sscanf(args[0], "%d", &ttl)
				if err != nil {
					return nil, fmt.Errorf("invalid TTL value: %s", args[0])
				}
				k8sCross.TTL = ttl
			case "cluster":
				// Parse cluster name
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, fmt.Errorf("cluster requires exactly one argument")
				}
				k8sCross.ClusterName = args[0]
			case "clusterset":
				// Parse cluster set name
				args := c.RemainingArgs()
				if len(args) != 1 {
					return nil, fmt.Errorf("clusterset requires exactly one argument")
				}
				k8sCross.ClusterSet = args[0]
			default:
				return nil, fmt.Errorf("unknown property '%s'", c.Val())
			}
		}
	}

	// Validate the configuration
	if k8sCross.HeadscaleClient == nil {
		return nil, fmt.Errorf("headscale_url is required configuration for k8s_cross plugin")
	}

	return k8sCross, nil
}
