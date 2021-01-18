package matchrelay

import (
	"net"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/infobloxopen/go-trees/iptree"
)

const pluginName = "matchrelay"

func init() { plugin.Register(pluginName, setup) }

func newDefaultFilter() *iptree.Tree {
	defaultFilter := iptree.NewTree()
	_, IPv4All, _ := net.ParseCIDR("0.0.0.0/0")
	_, IPv6All, _ := net.ParseCIDR("::/0")
	defaultFilter.InplaceInsertNet(IPv4All, struct{}{})
	defaultFilter.InplaceInsertNet(IPv6All, struct{}{})
	return defaultFilter
}

func setup(c *caddy.Controller) error {
	mr, err := parse(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		mr.Next = next
		return mr
	})

	return nil
}

func parse(c *caddy.Controller) (MatchRelay, error) {
	mr := New()
	for c.Next() {
		r := rule{}
		r.zones = c.RemainingArgs()
		if len(r.zones) == 0 {
			// if empty, the zones from the configuration block are used.
			r.zones = make([]string, len(c.ServerBlockKeys))
			copy(r.zones, c.ServerBlockKeys)
		}
		for i := range r.zones {
			r.zones[i] = plugin.Host(r.zones[i]).Normalize()
		}
		for c.NextBlock() {
			p := policy{}

			id := strings.ToLower(c.Val())
			if id == "net" {
				p.ftype = id
				p.filter = iptree.NewTree()
			} else if id != "relay" {
				return mr, c.Errf("unexpected token %q; expect 'net' or 'relay'", c.Val())
			}

			remainingTokens := c.RemainingArgs()
			if len(remainingTokens) == 0 {
				return mr, c.Errf("empty token")
			}
			if id == "net" {
				token := strings.ToLower(remainingTokens[0])
				if token == "*" {
					p.filter = newDefaultFilter()
					break
				}
				token = normalize(token)
				_, source, err := net.ParseCIDR(token)
				if err != nil {
					return mr, c.Errf("illegal CIDR notation %q", token)
				}
				p.filter.InplaceInsertNet(source, struct{}{})
			} else {
				for len(remainingTokens) > 0 {
					i := 0
					for ; i < len(remainingTokens) ; i++ {
						token := strings.ToLower(remainingTokens[i])
						mr.SetProxy(token)
					}
					remainingTokens = remainingTokens[i:]
				}
			}
			r.policies = append(r.policies, p)
		}
		mr.Rules = append(mr.Rules, r)
	}
	return mr, nil
}

// normalize appends '/32' for any single IPv4 address and '/128' for IPv6.
func normalize(rawNet string) string {
	if idx := strings.IndexAny(rawNet, "/"); idx >= 0 {
		return rawNet
	}

	if idx := strings.IndexAny(rawNet, ":"); idx >= 0 {
		return rawNet + "/128"
	}
	return rawNet + "/32"
}
