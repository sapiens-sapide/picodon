package explorators

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Backend interface {
	CreateInstanceIfNotExist(i Instance) error
	CreateAccountIfNotExist(a Account) error
	SaveAccount(Account) error
	SaveInstance(Instance) error
	FindAccountsToScan(inst *Instance) (accts []Account, err error)
	FindInstancesToScan() (instances []Instance, err error)
}

type PsqlDB struct {
	DB *gorm.DB
}

func (p *PsqlDB) CreateAccountIfNotExist(a Account) error {
	p.DB.FirstOrCreate(&Account{}, a)
	return nil
}

func (p *PsqlDB) CreateInstanceIfNotExist(i Instance) error {
	p.DB.FirstOrCreate(&Instance{}, i)
	return nil
}

// update account
func (p *PsqlDB) SaveAccount(a Account) error {
	var acct Account
	p.DB.Where("id = ? AND instance = ?", a.ID, a.Instance).FirstOrCreate(&acct)
	p.DB.Model(&acct).Save(a)
	return nil
}

// update instance
func (p *PsqlDB) SaveInstance(i Instance) error {
	var inst Instance
	p.DB.Where("domain = ?", i.Domain).FirstOrCreate(&inst)
	p.DB.Model(&inst).Save(i)
	return nil
}

func (p *PsqlDB) FindAccountsToScan(inst *Instance) (accts []Account, err error) {
	aWeekAgo := time.Now().Add(-(7 * 24 * time.Hour))
	p.DB.Where("last_scan isnull OR last_scan < ? and instance = ?", aWeekAgo, inst.Domain).Find(&accts)
	return
}

func (p *PsqlDB) FindInstancesToScan() (instances []Instance, err error) {
	aDayAgo := time.Now().Add(-24 * time.Hour)
	p.DB.Where("last_count isnull OR last_count < ? OR count_failed = true OR users_count = 0", aDayAgo).Find(&instances)
	return
}
