package abft

import (
	"github.com/Fantom-foundation/lachesis-base/abft/dagidx"
	"github.com/Fantom-foundation/lachesis-base/abft/election"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

type EpochDBLoadedFn func(idx.Epoch)

type OrdererDagIndex interface {
	dagidx.ForklessCause
}

// Orderer processes events to reach finality on their order.
// Unlike abft.Lachesis, this raw level of abstraction doesn't track cheaters detection
type Orderer struct {
	crit  func(error)
	store *Store
	input EventSource

	election *election.Election
	dagIndex OrdererDagIndex

	epochDBLoadedFn EpochDBLoadedFn
}

// NewOrderer creates Orderer instance.
// Unlike Lachesis, Orderer doesn't updates DAG indexes for events, and doesn't detect cheaters
// It has only one purpose - reaching consensus on events order.
func NewOrderer(store *Store, input EventSource, dagIndex OrdererDagIndex, crit func(error)) *Orderer {
	p := &Orderer{
		store:    store,
		input:    input,
		crit:     crit,
		dagIndex: dagIndex,
	}

	return p
}
