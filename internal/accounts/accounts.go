package accounts

import (
	"sort"
	"sync"
)

// accounts and balances. key: account, value: balance
var accounts map[string]int = make(map[string]int)
var lock sync.Locker = &sync.Mutex{}

// process deposit
func DepositMoney(account string, amount int) {
	println("DEPOSIT to", account, amount)
	lock.Lock()
	defer lock.Unlock()
	accounts[account] += amount
}

// process transfer
func TransferMoney(from string, to string, amount int) {
	lock.Lock()
	defer lock.Unlock()
	if accounts[from] < amount {
		return
	}
	println("TRANSFER", from, "->", to, amount)
	accounts[from] -= amount
	accounts[to] += amount
}

// show the accounts
func ShowAccounts() {
	lock.Lock()
	defer lock.Unlock()
	keys := make([]string, 0, len(accounts))
	for k := range accounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	print("BALANCES ")
	for _, k := range keys {
		if accounts[k] != 0 {
			print(k, ":", accounts[k], " ")
		}
	}
	println()
}
