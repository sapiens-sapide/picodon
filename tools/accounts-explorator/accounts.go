package accounts_explorator

import (
	"errors"
	"strings"
)

const addrSep = "@"

func splitUserAndInstance(acct, localInstance string) (user, instance string, err error) {
	switch strings.Count(acct, addrSep) {
	case 0:
		//a local user
		user = acct
		instance = localInstance
		return
	case 1:
		s := strings.Split(acct, addrSep)
		user = s[0]
		instance = s[1]
		return
	default:
		err = errors.New("invalid string")
		return
	}
}
