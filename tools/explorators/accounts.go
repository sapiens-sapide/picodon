package explorators

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const addrSep = "@"

func SplitUserAndInstance(acct, localInstance string) (user, instance string, err error) {
	switch strings.Count(acct, addrSep) {
	case 0:
		//a local user
		user = acct
		instance = localInstance
		return
	case 1:
		s := strings.Split(acct, addrSep)
		user = s[0]
		instance = normalizeInstanceDomain(instance)
		return
	default:
		err = errors.New("invalid string")
		return
	}
}

func GetRemoteAccountID(username, instance string) (id int, err error) {
	const webfingerService = ".well-known/webfinger"
	u := url.URL{
		Scheme:   "https",
		Host:     instance,
		Path:     webfingerService,
		RawQuery: "resource=acct%3A" + username + "%40" + instance,
	}
	resp, err := http.Get(u.String())
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var payload webfinger
	json.Unmarshal(body, &payload)
	var p []string
	for _, link := range payload.Links {
		if link.Rel == "salmon" {
			u, err := url.Parse(link.Href)
			if err != nil {
				return 0, err
			}
			p = strings.Split(u.Path, "/")
		}
	}
	l := len(p)
	if l == 4 && p[l-1] != "" {
		id, err = strconv.Atoi(p[l-1])
		if err != nil {
			return 0, err
		}
		return
	} else {
		return 0, errors.New("empty response")
	}

}

type webfinger struct {
	Subject string
	Aliases []string
	Links   []Link
}

type Link struct {
	Rel  string
	Type string
	Href string
}

func normalizeInstanceDomain(domain string) (d string) {
	d = strings.ToLower(domain)
	d = strings.TrimSuffix(d, ".")
	d = strings.TrimSuffix(d, "/")
	return
}
