// Package matchrelay implements a plugin that match source ip and relay to upstream
package matchrelay

import (
	"context"
	"net"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/request"

	"github.com/infobloxopen/go-trees/iptree"
	"github.com/miekg/dns"
)

// MatchRelay is a plugin that matches your IP address used for connecting to CoreDNS.
type MatchRelay struct{
	Next		plugin.Handler

	fwd		*forward.Forward
	rules		[]rule
	zones		[]string
	interval	time.Duration
	filename	string
}

type rule struct {
	policies	[]policy
}

type policy struct {
	ftype	string
	filter	*iptree.Tree
}

func New() MatchRelay {
	mr := MatchRelay{}
	mr.fwd = forward.New()
	return mr
}

func (mr MatchRelay) SetProxy(proxy string) {
	mr.fwd.SetProxy(forward.NewProxy(proxy, "dns"))
}

// ServeDNS implements the plugin.Handler interface.
func (mr *MatchRelay) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	for _, rule := range mr.rules {
		// check zone.
		zone := plugin.Zones(mr.zones).Matches(state.Name())
		if zone == "" {
			continue
		}

		ipMatch := matchWithPolicies(rule.policies, w, r)
		if ipMatch {
			mr.fwd.ServeDNS(ctx, w, r)

			return 0, nil
		}
	}
	return plugin.NextOrFailure(state.Name(), mr.Next, ctx, w, r)
}

// matchWithPolicies matches the DNS query with a list of Match polices and returns boolean
func matchWithPolicies(policies []policy, w dns.ResponseWriter, r *dns.Msg) bool {
	state := request.Request{W: w, Req: r}

	ip := net.ParseIP(state.IP())
	for _, policy := range policies {
		_, contained := policy.filter.GetByIP(ip)
		if !contained {
			continue
		}

		// matched.
		return true
	}
	return false
}

// Name implements the Handler interface.
func (mr MatchRelay) Name() string { return pluginName }
