package matchrelay

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func TestMatchRelay(t *testing.T) {
	mr := MatchRelay{}
	if mr.Name() != pluginName {
		t.Errorf("expected plugin name: %s, got %s", mr.Name(), pluginName)
	}
}
