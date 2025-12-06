# k8s_cross

## Name

*k8s_cross* - Kubernetes cross-cluster DNS resolution plugin.

## Description

The *k8s_cross* plugin provides cross-cluster service discovery for Kubernetes using the Multi-Cluster Services API (KEP-1645) standard. It integrates with Headscale to enable DNS resolution for services across multiple Kubernetes clusters, allowing applications to transparently discover and connect to services running in different clusters using standard DNS queries.

The plugin follows the KEP-1645 Multi-Cluster Services API specifications, supporting the `clusterset.local` domain for cross-cluster service discovery. It enables ServiceExport and ServiceImport patterns by leveraging Headscale's secure network overlay capabilities.

## Syntax

```
k8s_cross [ZONE...] {
    headscale_url URL API_KEY
    cluster CLUSTER_NAME
    clusterset CLUSTERSET_NAME
    ttl TTL_VALUE
}
```

* **ZONE**: The zones for which the plugin is authoritative. If not specified, defaults to "." (all zones).
* `headscale_url`: Specifies the URL and API key for the Headscale server. Required.
* `cluster`: Specifies the local cluster name. Defaults to "default-cluster".
* `clusterset`: Specifies the clusterset name. Defaults to "default-clusterset".
* `ttl`: Sets the TTL value for DNS records. Defaults to 300 seconds.

## Examples

### Basic Configuration

```
clusterset.local {
    k8s_cross {
        headscale_url http://headscale:8080 your-headscale-api-key
        cluster prod-cluster
        clusterset production
        ttl 600
    }
}
```

This configuration enables cross-cluster DNS resolution for the `clusterset.local` domain, connecting to the Headscale server at `http://headscale:8080` using the provided API key. DNS records will have a TTL of 600 seconds.

### Multi-Zone Configuration

```
example.org {
    k8s_cross clusterset.local {
        headscale_url https://headscale.example.com:443 api-key-12345
        cluster primary-cluster
        clusterset staging
    }
}
```

## KEP-1645 Compliance

The plugin implements the Multi-Cluster Services API (KEP-1645) by:

1. Using the standard `clusterset.local` domain for cross-cluster service discovery
2. Supporting service queries in the format: `<service>.<namespace>.svc.clusterset.local`
3. Providing DNS records (A, AAAA, SRV, TXT) for multi-cluster services
4. Supporting both ClusterSetIP and headless service patterns

## Metrics

The plugin exports the following Prometheus metrics:

* `coredns_k8s_cross_request_count_total{server}`: Counter of DNS requests processed by the k8s_cross plugin.

## Ready

The plugin does not implement the Ready interface, as it's always ready to process requests after initialization.

## See also

* [KEP-1645: Multi-Cluster Services API](https://github.com/kubernetes/enhancements/tree/master/keps/sig-multicluster/1645-multi-cluster-services-api)
* [Headscale Documentation](https://headscale.net/)
* [CoreDNS Documentation](https://coredns.io/)