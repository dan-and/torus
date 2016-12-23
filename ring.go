package torus

import (
	"math/big"

	"github.com/coreos/torus/models"
)

type RingType int

type Ring interface {
	GetPeers(key BlockRef) (PeerPermutation, error)
	Members() PeerList

	Describe() string
	Type() RingType
	Version() int

	Marshal() ([]byte, error)
}

type ModifyableRing interface {
	ChangeReplication(r int) (Ring, error)
}

type RingAdder interface {
	ModifyableRing
	AddPeers(PeerInfoList) (Ring, error)
}

type RingRemover interface {
	ModifyableRing
	RemovePeers(PeerList) (Ring, error)
}

type PeerPermutation struct {
	Replication int
	Peers       PeerList
}

type PeerList []string

func (pl PeerList) IndexAt(uuid string) int {
	for i, x := range pl {
		if x == uuid {
			return i
		}
	}
	return -1
}

func (pl PeerList) Has(uuid string) bool {
	return pl.IndexAt(uuid) != -1
}

func (pl PeerList) AndNot(b PeerList) PeerList {
	var out PeerList
	for _, x := range pl {
		if !b.Has(x) {
			out = append(out, x)
		}
	}
	return out
}

func (pl PeerList) Union(b PeerList) PeerList {
	var out PeerList
	for _, x := range pl {
		out = append(out, x)
	}
	for _, x := range b {
		if !pl.Has(x) {
			out = append(out, x)
		}
	}
	return out
}

func (pl PeerList) Intersect(b PeerList) PeerList {
	var out PeerList
	for _, x := range pl {
		if b.Has(x) {
			out = append(out, x)
		}
	}
	return out
}

// Applicative! Applicative! My kingdom for Applicative!

type PeerInfoList []*models.PeerInfo

func (pi PeerInfoList) UUIDAt(uuid string) int {
	for i, x := range pi {
		if x.UUID == uuid {
			return i
		}
	}
	return -1
}

func (pi PeerInfoList) HasUUID(uuid string) bool {
	return pi.UUIDAt(uuid) != -1
}

func (pi PeerInfoList) AndNot(b PeerList) PeerInfoList {
	var out PeerInfoList
	for _, x := range pi {
		if !b.Has(x.UUID) {
			out = append(out, x)
		}
	}
	return out
}

func (pi PeerInfoList) Union(b PeerInfoList) PeerInfoList {
	var out PeerInfoList
	for _, x := range pi {
		out = append(out, x)
	}
	for _, x := range b {
		if !pi.HasUUID(x.UUID) {
			out = append(out, x)
		}
	}
	return out
}

func (pi PeerInfoList) Intersect(b PeerInfoList) PeerInfoList {
	var out PeerInfoList
	for _, x := range pi {
		if b.HasUUID(x.UUID) {
			out = append(out, x)
		}
	}
	return out
}

func (pi PeerInfoList) PeerList() PeerList {
	out := make([]string, len(pi))
	for i, x := range pi {
		out[i] = x.UUID
	}
	return PeerList(out)
}

func (pi PeerInfoList) GetWeights() map[string]int {
	out := make(map[string]int)
	if len(pi) == 0 {
		return out
	}
	gcd := big.NewInt(int64(pi[0].TotalBlocks))
	for _, p := range pi[1:] {
		gcd.GCD(nil, nil, gcd, big.NewInt(int64(p.TotalBlocks)))
	}
	for _, p := range pi {
		out[p.UUID] = int(p.TotalBlocks / uint64(gcd.Int64()))
		clog.Infof("%s: weight %d", p.UUID, out[p.UUID])
	}
	return out
}
