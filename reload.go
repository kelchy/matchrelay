package matchrelay

import (
	"io/ioutil"
	"strings"

	"github.com/coredns/coredns/plugin/pkg/log"
)

// Reload - function which reloads the rules
func (mr *MatchRelay) Reload(buf []byte) {
	mr.rules = nil
	mr.domains = make(map[string]string)
	lines := strings.Split(string(buf), "\n")
	r := rule{}
	for _, line := range lines {
		fields := strings.Split(line, " ")
		log.Infof("read line %s - type %s\n", line, fields[0])
		if  fields[0] == "net" {
			id := fields[0]
			fields = fields[1:]
			p := makePolicy(fields)
			if p.filter != nil {
				p.ftype = id
				r.policies = append(r.policies, p)

			}
		} else if fields[0] == "domain" {
			if fields[1] != "" {
				log.Infof("%s adding key %s to domain\n", fields[1])
				mr.domains[fields[1]] = ""
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
