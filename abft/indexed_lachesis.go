package abft

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
)

var _ lachesis.Consensus = (*IndexedLachesis)(nil)

// IndexedLachesis performs events ordering and detects cheaters
// It's a wrapper around Orderer, which adds features which might potentially be application-specific:
// confirmed events traversal, DAG index updates and cheaters detection.
// Use this structure if need a general-purpose consensus. Instead, use lower-level abft.Orderer.
type IndexedLachesis struct {
	*Lachesis
	uniqueDirtyID uniqueID
}

type DagIndexer interface {
	DagIndex

	Add(dag.Event) error
	Flush()
	DropNotFlushed()

	Reset(validators *pos.Validators, db kvdb.Store, getEvent func(hash.Event) dag.Event)
}

// New creates IndexedLachesis instance.
func NewIndexedLachesis(store *Store, input EventSource, dagIndexer DagIndexer, crit func(error)) *IndexedLachesis {
	p := &IndexedLachesis{
		Lachesis:      NewLachesis(store, input, dagIndexer, crit),
		uniqueDirtyID: uniqueID{new(big.Int)},
	}

	return p
}

// Build fills consensus-related fields: Frame, IsRoot
// returns error if event should be dropped
func (p *IndexedLachesis) Build(e dag.MutableEvent) error {
	dagIndex := p.dagIndex.(DagIndexer)

	e.SetID(p.uniqueDirtyID.sample())

	defer dagIndex.DropNotFlushed()
	err := dagIndex.Add(e)
	if err != nil {
		return err
	}

	return p.Lachesis.Build(e)
}

// Process takes event into processing.
// Event order matter: parents first.
// All the event checkers must be launched.
// Process is not safe for concurrent use.
func (p *IndexedLachesis) Process(e dag.Event) (err error) {
	dagIndex := p.dagIndex.(DagIndexer)

	defer dagIndex.DropNotFlushed()
	err = dagIndex.Add(e)
	if err != nil {
		return err
	}

	err = p.Lachesis.Process(e)
	if err != nil {
		return err
	}
	dagIndex.Flush()
	return nil
}

func (p *IndexedLachesis) Bootstrap(beginBlockFn lachesis.BeginBlockFn) error {
	epochDBloadedFn := func(epoch idx.Epoch) {
		p.dagIndex.(DagIndexer).Reset(p.store.GetValidators(), p.store.epochTable.VectorIndex, p.input.GetEvent)
	}
	return p.Lachesis.BootstrapWithOrderer(beginBlockFn, epochDBloadedFn)
}

type uniqueID struct {
	counter *big.Int
}

func (u *uniqueID) sample() [24]byte {
	u.counter = u.counter.Add(u.counter, common.Big1)
	var id [24]byte
	copy(id[:], u.counter.Bytes())
	return id
}
