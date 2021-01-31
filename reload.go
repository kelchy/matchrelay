package matchrelay

import (
	"io/ioutil"
	"strings"

	"github.com/coredns/coredns/plugin/pkg/log"
)

func (mr *MatchRelay) Reload(buf []byte) {
	mr.rules = nil
	lines := strings.Split(string(buf), "\n")
	r := rule{}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		if  fields[0] == "net" {
			id := fields[0]
			fields = fields[1:]
			p := makePolicy(fields)
			if p.filter != nil {
				p.ftype = id
				r.policies = append(r.policies, p)

			}
		}
	}
	if len(r.policies) > 0 {
		mr.rules = append(mr.rules, r)
	}
}

func fileOpen(fileName string) ([]byte, error) {

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Errorf("error opening file %s", fileName)
		return file, err
	}
	return file, nil
}