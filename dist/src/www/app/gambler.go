package app

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	Users        = make(map[string]U)
	MAX_BALANCES = 6
)

type U struct {
	Uname string `json:"uname"`

	// total balance the user owns
	TotalBalance uint64 `json:"total_balance"`
	balances     []uint64
}

func (u *U) CalcBalance() uint64 {
	var t uint64
	for i, _ := range u.balances {
		t += u.balances[i]
	}
	return t
}

func (u *U) Reset() {
	u.balances = make([]uint64, 0)
	u.TotalBalance = 0
}

func (u *U) Beg() {
	if len(u.balances) >= MAX_BALANCES {
		return
	}
	b := rand.Intn(99)
	u.balances = append(u.balances, uint64(b))
}

func NewUser(uname string) U {
	u := U{Uname: uname}
	u.Reset()
	return u
}
