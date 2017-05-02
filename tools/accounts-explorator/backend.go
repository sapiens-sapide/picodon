package accounts_explorator

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Backend interface {
	SaveAccount(Account) error
	SaveInstance(Instance) error
	FindAccountsToScan(inst *Instance) (accts []Account, err error)
}

type PsqlDB struct {
	DB *gorm.DB
}

// create or update account
func (p *PsqlDB) SaveAccount(a Account) error {
	p.DB.Where("id = ? AND instance = ?", a.ID, a.Instance).Assign(a).FirstOrCreate(&a)
	return nil
}

// create or update instance
func (p *PsqlDB) SaveInstance(i Instance) error {
	p.DB.Where("domain = ?", i.Domain).Assign(i).FirstOrCreate(&i)
	return nil
}

func (p *PsqlDB) FindAccountsToScan(inst *Instance) (accts []Account, err error) {
	aWeekAgo := time.Now().Add(-(7 * 24 * time.Hour))
	p.DB.Where("last_scan isnull OR last_scan < ? and instance = ?", aWeekAgo, inst.Domain).Find(&accts)
	return
}