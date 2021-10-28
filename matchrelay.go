// Package matchrelay implements a plugin that match source ip and relay to upstream
package matchrelay

import (
	"context"
	"net"
	"time"
	"strings"
	"crypto/md5"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/forward"
	"github.com/coredns/coredns/request"
	"github.com/coredns/coredns/plugin/pkg/log"

	"github.com/infobloxopen/go-trees/iptree"
	"github.com/miekg/dns"
)

// MatchRelay is a plugin that matches your IP address used for connecting to CoreDNS.
type MatchRelay struct{
	Next		plugin.Handler

	fwd		*forward.Forward
	rules		[]rule
	zones		[]string
	domains		map[string]string
	interval	time.Duration
	files		[]string
	md5sum		map[string][16]byte
}

type rule struct {
	policies	[]policy
}

type policy struct {
	ftype	string
	filter	*iptree.Tree
}

// New - function which creates a module instance on coredns
func New() MatchRelay {
	mr := MatchRelay{}
	mr.fwd = forward.New()
	return mr
}

// SetProxy - function which sets forwarding relay
func (mr MatchRelay) SetProxy(proxy string) {
	mr.fwd.SetProxy(forward.NewProxy(proxy, "dns"))
}

// ServeDNS - function which implements the plugin.Handler interface.
func (mr *MatchRelay) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	if len(mr.domains) > 0 {
		sArr := strings.Split(state.Name(), ".")
		if len(sArr) > 0 {
			// state.Name() will always have a trailing .
			// remove last element
			sArr = sArr[:len(sArr)-1]
		}
		base := sArr[len(sArr) - 1]
		for i := len(sArr) - 2; i >= 0; i = i - 1 {
			str := sArr[i] + "."  + base
			log.Infof("Matchrelay matching %s from %d entries\n", str, len(mr.domains))
			if _, ok := mr.domains[str]; ok {
				mr.fwd.ServeDNS(ctx, w, r)
				return 0, nil
			}
			base = str
		}
		log.Infof("Matchrelay no match %s\n", state.Name())
		return plugin.NextOrFailure(state.Name(), mr.Next, ctx, w, r)
	}

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

func (mr *MatchRelay) pushMatch() error {
	var buf []byte
	changed := false
	for _, file := range mr.files {
		s, e := fileOpen(file)
		if e != nil {
			log.Errorf("pushMatch error opening matchrelay file %s", file)
			return e
		}
		md5sum := md5.Sum(s)
		if  mr.md5sum[file] != md5sum {
			log.Infof("Matchrelay new config %s MD5 = %x\n", file, md5sum)
			changed = true
			mr.md5sum[file] = md5sum
		}
		// insert a new line character (10) in between files just to be sure
		buf = append(buf, append(s, 10)...)
	}
	if changed {
		log.Infof("batch files %s", string(buf))
		mr.Reload(buf)
	}
	return nil
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
