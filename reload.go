package matchrelay

import (
	"io/ioutil"
	"strings"
	"strconv"
	"bufio"
	"bytes"

	"github.com/coredns/coredns/plugin/pkg/log"
)

// Reload - function which reloads the rules
func (mr *MatchRelay) Reload(buf []byte) {
	mr.rules = nil
	mr.domains = make(map[string]string)

	r := rule{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
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
				mr.domains[fields[1]] = strconv.Itoa(len(buf))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("read line error %v", err)
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
