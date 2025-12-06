// Package k8s_cross is a CoreDNS plugin that implements KEP-1645 Multi-Cluster Services API
// functionality using Headscale for cross-cluster connectivity.
//
// The plugin provides DNS resolution for services across multiple Kubernetes clusters
// following the ServiceExport and ServiceImport patterns.
package k8s_cross

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"

	"github.com/wold9168/k8s_cross/headscale"
)

// Define a logger with the plugin name. This allows us to use log.Info and
// other related methods for logging.
var log = clog.NewWithPlugin("k8s_cross")

// K8sCross is the main structure for the k8s_cross plugin, handling DNS requests for multi-cluster services.
type K8sCross struct {
	Next plugin.Handler

	// Configuration for the plugin
	HeadscaleClient headscale.HeadscaleClient
	Zones           []string
	TTL             uint32
	ClusterName     string
	ClusterSet      string
}

// ServeDNS implements the plugin.Handler interface. This is the entry point for the plugin to handle DNS requests.
// Parameters:
// - ctx: Request context containing request-related information
// - w: DNS response writer used to send responses to clients
// - r: DNS request message
// Returns:
// - int: DNS response code
// - error: Error during processing
func (e K8sCross) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Log received request for debugging
	log.Debug("Received DNS request")

	// Parse the incoming request
	state := request.Request{W: w, Req: r}
	q := state.Req.Question[0]

	// Check if this is a service resolution request for multi-cluster services
	// Pattern: <service>.<namespace>.svc.clusterset.local
	qName := strings.ToLower(q.Name)

	// Check if the query matches the clusterset.local domain
	if !e.isClustersetQuery(qName) {
		// If not a clusterset query, pass to next handler in chain
		return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
	}

	// Increase request count metric
	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	log.Debugf("Processing clusterset.local query: %s", qName)

	// Handle the multi-cluster service query
	resp, err := e.handleClustersetQuery(ctx, state, q)
	if err != nil {
		log.Errorf("Error handling clusterset query: %v", err)
		return dns.RcodeServerFailure, err
	}

	// Write the response
	err = w.WriteMsg(resp)
	if err != nil {
		log.Errorf("Error writing response: %v", err)
		return dns.RcodeServerFailure, err
	}

	return dns.RcodeSuccess, nil
}

// isClustersetQuery checks if the DNS query targets the clusterset.local domain
func (e K8sCross) isClustersetQuery(name string) bool {
	for _, zone := range e.Zones {
		if strings.HasSuffix(name, "."+zone+".") {
			return true
		}
	}
	return false
}

// handleClustersetQuery handles DNS queries for services in the clusterset.local domain
func (e K8sCross) handleClustersetQuery(ctx context.Context, state request.Request, q dns.Question) (*dns.Msg, error) {
	resp := new(dns.Msg)
	resp.SetReply(state.Req)
	resp.Authoritative = true

	qName := strings.ToLower(q.Name)
	qType := q.Qtype

	// Parse the domain name to extract service, namespace, and cluster information
	service, namespace, isValid := e.parseClusterSetDomain(qName)
	if !isValid {
		log.Debugf("Invalid clusterset domain: %s", qName)
		resp.SetRcode(state.Req, dns.RcodeNameError)
		return resp, nil
	}

	log.Debugf("Processing query for service: %s, namespace: %s, type: %s", service, namespace, dns.TypeToString[qType])

	// Find nodes that match the service and namespace in the Headscale network
	nodes, err := e.findServiceNodes(ctx, service, namespace)
	if err != nil {
		log.Errorf("Error finding service nodes: %v", err)
		resp.SetRcode(state.Req, dns.RcodeServerFailure)
		return resp, nil
	}

	// Build DNS records based on the found nodes
	var answers []dns.RR
	switch qType {
	case dns.TypeA:
		answers = e.buildARecords(nodes, service, namespace)
	case dns.TypeAAAA:
		answers = e.buildAAAARecords(nodes, service, namespace)
	case dns.TypeSRV:
		answers = e.buildSRVRecords(nodes, service, namespace)
	case dns.TypeTXT:
		answers = e.buildTXTRecords(nodes, service, namespace)
	default:
		// For unsupported types, just return no error
		resp.SetRcode(state.Req, dns.RcodeSuccess)
		return resp, nil
	}

	resp.Answer = answers
	return resp, nil
}

// parseClusterSetDomain parses a clusterset.local domain and extracts service and namespace
func (e K8sCross) parseClusterSetDomain(name string) (service, namespace string, valid bool) {
	name = strings.TrimSuffix(name, ".")
	
	// Expected format: <service>.<namespace>.svc.clusterset.local
	// Example: my-service.my-namespace.svc.clusterset.local
	parts := strings.Split(name, ".")
	
	if len(parts) < 5 {
		return "", "", false
	}
	
	// Check if domain ends with "svc.clusterset.local"
	if parts[len(parts)-1] != "local" || parts[len(parts)-2] != "clusterset" || parts[len(parts)-3] != "svc" {
		return "", "", false
	}
	
	// Extract namespace and service
	if len(parts) >= 5 {
		namespace = parts[len(parts)-4] // fourth from the end
		service = parts[len(parts)-5]    // fifth from the end
	}
	
	return service, namespace, true
}

// findServiceNodes queries Headscale to find nodes matching the service and namespace
func (e K8sCross) findServiceNodes(ctx context.Context, service, namespace string) ([]*headscale.Node, error) {
	// In a real implementation, this would query the Headscale API for nodes
	// that match the service and namespace. For now, we'll simulate this by
	// listing all nodes and filtering them.
	// 
	// In practice, you'd need to tag or label nodes in Headscale with service
	// and namespace information, then query by those properties.
	
	nodes, err := e.HeadscaleClient.ListNodes(ctx, "")
	if err != nil {
		return nil, err
	}

	// Filter nodes based on service and namespace (this would be done on the server side in a full implementation)
	var matchingNodes []*headscale.Node
	for i := range nodes {
		node := &nodes[i]
		// In a real implementation, you would filter based on node labels/tags that indicate
		// which service and namespace the node belongs to
		// For now, we check if the node name contains both service and namespace
		nodeName := strings.ToLower(node.Name)
		if strings.Contains(nodeName, strings.ToLower(service)) && strings.Contains(nodeName, strings.ToLower(namespace)) {
			matchingNodes = append(matchingNodes, node)
		}
	}

	return matchingNodes, nil
}

// buildARecords creates A records for the IP addresses of the nodes
func (e K8sCross) buildARecords(nodes []*headscale.Node, service, namespace string) []dns.RR {
	var records []dns.RR
	
	for _, node := range nodes {
		for _, ipStr := range node.IPAddresses {
			ip := net.ParseIP(ipStr)
			if ip != nil && ip.To4() != nil { // IPv4 only
				aRecord := &dns.A{
					Hdr: dns.RR_Header{
						Name:   fmt.Sprintf("%s.%s.svc.clusterset.local.", service, namespace),
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    e.TTL,
					},
					A: ip,
				}
				records = append(records, aRecord)
			}
		}
	}
	
	return records
}

// buildAAAARecords creates AAAA records for the IP addresses of the nodes
func (e K8sCross) buildAAAARecords(nodes []*headscale.Node, service, namespace string) []dns.RR {
	var records []dns.RR
	
	for _, node := range nodes {
		for _, ipStr := range node.IPAddresses {
			ip := net.ParseIP(ipStr)
			if ip != nil && ip.To4() == nil { // IPv6 only
				aaaaRecord := &dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   fmt.Sprintf("%s.%s.svc.clusterset.local.", service, namespace),
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    e.TTL,
					},
					AAAA: ip,
				}
				records = append(records, aaaaRecord)
			}
		}
	}
	
	return records
}

// buildSRVRecords creates SRV records for the service
func (e K8sCross) buildSRVRecords(nodes []*headscale.Node, service, namespace string) []dns.RR {
	var records []dns.RR

	// SRV records follow the format _service._proto.name. TTL class SRV priority weight port target
	// For the service, we create one SRV record regardless of the number of nodes
	if len(nodes) > 0 {
		srvRecord := &dns.SRV{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("_http._tcp.%s.%s.svc.clusterset.local.", service, namespace),
				Rrtype: dns.TypeSRV,
				Class:  dns.ClassINET,
				Ttl:    e.TTL,
			},
			Priority: 10,
			Weight:   10,
			Port:     80,
			Target:   fmt.Sprintf("%s.%s.svc.clusterset.local.", service, namespace),
		}
		records = append(records, srvRecord)
	}

	return records
}

// buildTXTRecords creates TXT records for the service
func (e K8sCross) buildTXTRecords(nodes []*headscale.Node, service, namespace string) []dns.RR {
	var records []dns.RR

	if len(nodes) > 0 {
		txtRecord := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   fmt.Sprintf("%s.%s.svc.clusterset.local.", service, namespace),
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    e.TTL,
			},
			Txt: []string{
				fmt.Sprintf("cluster=%s", e.ClusterName),
				fmt.Sprintf("clusterset=%s", e.ClusterSet),
				fmt.Sprintf("service=%s", service),
				fmt.Sprintf("namespace=%s", namespace),
			},
		}
		records = append(records, txtRecord)
	}

	return records
}

// Name implements the Handler interface, returning the plugin name.
func (e K8sCross) Name() string { return "k8s_cross" }