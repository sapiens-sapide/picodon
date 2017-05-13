package instancesWorker

import (
	"context"
	"fmt"
	"github.com/sapiens-sapide/go-mastodon"
	expl "github.com/sapiens-sapide/picodon/tools/explorators"
	"sync"
)

type InstanceWorker struct {
	Backend       expl.Backend
	Context       context.Context
	Instance      expl.Instance
	IsWSConnected bool        // whether worker has a websocket connection established with instance
	SeenLock      *sync.Mutex // to prevent concurrent access to seen maps below
	AccountsSeen  map[string]bool
	InstancesSeen map[string]bool
}

func (iw *InstanceWorker) SaveIfUnknown(acc mastodon.Account) (acct expl.Account, NewAccount, NewInstance bool) {
	user, instance, err := expl.SplitUserAndInstance(acc.Acct, iw.Instance.Domain)
	if err != nil {
		fmt.Printf("error : %s\n", err)
		return
	}
	acct = expl.Account{
		Username: user,
		Instance: instance,
	}
	iw.SeenLock.Lock()
	if _, ok := iw.AccountsSeen[acct.Username+"@"+acct.Instance]; !ok {
		if instance != iw.Instance.Domain {
			id, err := expl.GetRemoteAccountID(user, instance)
			if err == nil {
				acct.ID = uint(id)
			}
		} else {
			acct.ID = uint(acc.ID)
		}
		if acct.ID != 0 {
			iw.Backend.CreateAccountIfNotExist(acct)
		}
		if _, ok := iw.InstancesSeen[acct.Instance]; !ok {
			iw.Backend.CreateInstanceIfNotExist(expl.Instance{Domain: instance})
			iw.InstancesSeen[acct.Instance] = true
		}
		iw.AccountsSeen[acct.Username+"@"+acct.Instance] = true
	}
	iw.SeenLock.Unlock()
	//TODO: feedback newaccount&newinstance
	return
}
