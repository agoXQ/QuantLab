package tushare

import (
	domainProvider "github.com/agoXQ/QuantLab/app/market/domain/provider"
)

// Provider is the Tushare-backed implementation of domain.DataProvider.
//
// The provider is composed of small files (security.go, bars.go, ...), each
// adding a fetcher method to the same struct. This keeps the implementation
// readable while still satisfying the wide DataProvider interface.
type Provider struct {
	client *Client
}

// NewProvider returns a Tushare-backed DataProvider.
func NewProvider(client *Client) domainProvider.DataProvider {
	return &Provider{client: client}
}

// Name implements domain.DataProvider.
func (p *Provider) Name() string { return "tushare" }
