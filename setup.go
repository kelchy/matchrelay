package matchrelay

import (
	"net"
	"strings"
	"time"
	"path/filepath"
	"crypto/md5"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"

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

	loop := make(chan bool)
	c.OnStartup(func() error {
		if mr.interval == 0 || mr.filename == "" {
			return nil
		}
		s, e := fileOpen(mr.filename)
		if e != nil {
			log.Errorf("error opening matchrelay file %s", mr.filename)
			return e
		}
		md5sum := md5.Sum(s)
		mr.Reload(s)

		go func() {
			ticker := time.NewTicker(mr.interval)
			for {
				select {
				case <-loop:
					return
				case <-ticker.C:
					s, e := fileOpen(mr.filename)
					if e != nil {
						log.Errorf("error opening matchrelay file %s", mr.filename)
						return
					}
					ms := md5.Sum(s)
					if md5sum != ms {
						log.Infof("Matchrelay new config MD5 = %x\n", ms)
						md5sum = ms
						mr.Reload(s)
					}
				}
			}
		}()
		return nil
	})

	c.OnShutdown(func() error {
		close(loop)
		return nil
	})

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		mr.Next = next
		return &mr
	})

	return nil
}

func parse(c *caddy.Controller) (MatchRelay, error) {
	mr := New()
	// matchrelay takes zone details from server block, not on config block
	mr.zones = make([]string, len(c.ServerBlockKeys))
	mr.domains = make(map[string]string)
	copy(mr.zones, c.ServerBlockKeys)
	for i := range mr.zones {
		mr.zones[i] = plugin.Host(mr.zones[i]).Normalize()
	}
	for c.Next() {
		r := rule{}
		for c.NextBlock() {
			id := strings.ToLower(c.Val())
			remainingTokens := c.RemainingArgs()
			if len(remainingTokens) == 0 {
				return mr, c.Errf("empty token")
			}
			switch id {
			case "domain":
				// we don't support any options for now so set it to empty string
				mr.domains[remainingTokens[0]] = ""
			case "net":
				// static rules
				p := makePolicy(remainingTokens)
				if p.filter != nil {
					p.ftype = id
					r.policies = append(r.policies, p)
				}
			case "reload":
				// TODO: add jitter
				d, err := time.ParseDuration(remainingTokens[0])
				if err != nil {
					return mr, plugin.Error("invalid reload timer", err)
				}
				mr.interval = d
			case "relay":
				for len(remainingTokens) > 0 {
					i := 0
					for ; i < len(remainingTokens) ; i++ {
						token := strings.ToLower(remainingTokens[i])
						mr.SetProxy(token)
					}
					remainingTokens = remainingTokens[i:]
				}
			case "match":
				// file based rules with own reload mechanism compatible with static rules above
				fileName := strings.ToLower(remainingTokens[0])
				config := dnsserver.GetConfig(c)
				if !filepath.IsAbs(fileName) && config.Root != "" {
					fileName = filepath.Join(config.Root, fileName)
				}
				mr.filename = fileName
			default:
				return mr, c.Errf("unexpected token %q; expect 'net', 'match', 'reload' or 'relay'", id)
			}
		}
		if len(r.policies) > 0 {
			mr.rules = append(mr.rules, r)
		}
	}
	return mr, nil
}

// take the cidrs and build the policy
func makePolicy(rule []string) policy {
	p := policy{}

	// TODO: handle multiple CIDR, watch out for inline comments which may end up in rule slice
	token := strings.ToLower(rule[0])
	if token == "*" {
		p.filter = newDefaultFilter()
		return p
	}
	token = normalize(token)
	_, source, err := net.ParseCIDR(token)
	if err != nil {
		log.Errorf("illegal CIDR notation %q", token)
		return p
	}
	p.filter = iptree.NewTree()
	p.filter.InplaceInsertNet(source, struct{}{})
	return p
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
