package app

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

const MAX_USER = 10

type Service struct {
	// formal player has already been confirmed by administrator
	Players map[string]U

	// players who applies to join the game but has not been confirmed by admin
	Pendings map[string]U

	// winners of this round
	Winners map[string]struct{}

	// is service running?
	Doing bool

	// this round's duration
	Duration time.Duration
	m        sync.Mutex
}

var (
	// singleton
	CasinoService = NewService()
)

func NewService() *Service {
	return &Service{
		Players:  make(map[string]U),
		Pendings: make(map[string]U),
		Winners:  make(map[string]struct{}),
		Doing:    false,
		Duration: time.Minute * 5,
		m:        sync.Mutex{},
	}
}

// applies to join
func (s *Service) ApplicatUser(u U) {
	s.m.Lock()
	defer s.m.Unlock()
	for name, _ := range s.Pendings {
		if name == u.Uname {
			return
		}
	}
	for name, _ := range s.Players {
		if name == u.Uname {
			return
		}
	}

	s.Pendings[u.Uname] = u
}

// add user from []pendings to []players
func (s *Service) AddPlayer(u U) error {
	s.m.Lock()
	defer s.m.Unlock()

	if len(s.Players) >= MAX_USER {
		return errors.New("players are enough")
	}

	if _, ok := s.Players[u.Uname]; ok {
		return errors.New("you are already a player")
	}

	if _, ok := s.Pendings[u.Uname]; !ok {
		return errors.New("not in pending list")
	}
	delete(s.Pendings, u.Uname)
	s.Players[u.Uname] = u

	return nil
}

func sum(b []uint64) uint64 {
	var ret uint64
	for i, _ := range b {
		ret += b[i]
	}
	return ret
}

func (s *Service) Calc() {
	time.Sleep(s.Duration)

	s.m.Lock()
	defer s.m.Unlock()

	var total []uint64
	for _, player := range s.Players {
		// the way to win
		// e.g.
		// suppose your balances is [99, 99, 99, 99, 99, 99]
		// if 99*6 + $RANDOM == 0x1010010C then you win!
		// GOOD LUCK
		total = player.balances
		total = append(total, uint64(0xFFFFFFF+rand.Intn(0xFFFFF)))
		if sum(total) == 0x1010010C {
			s.Winners[player.Uname] = struct{}{}
		}
	}
	s.Doing = false
}

func (s *Service) Start() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.Doing {
		return
	}

	s.Doing = true

	go func() {
		s.Calc()
	}()

}

func (s *Service) Reset() {
	s.m.Lock()
	defer s.m.Unlock()
	if s.Doing {
		return
	}

	s.Players = make(map[string]U)
	s.Pendings = make(map[string]U)
	s.Winners = make(map[string]struct{})
}
