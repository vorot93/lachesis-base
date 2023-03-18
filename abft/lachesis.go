package abft

import (
	"errors"

	"github.com/Fantom-foundation/lachesis-base/abft/dagidx"
	"github.com/Fantom-foundation/lachesis-base/abft/election"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/Fantom-foundation/lachesis-base/lachesis"
)

var _ lachesis.Consensus = (*Lachesis)(nil)

type DagIndex interface {
	dagidx.VectorClock
	dagidx.ForklessCause
}

// Lachesis performs events ordering and detects cheaters
// It's a wrapper around Orderer, which adds features which might potentially be application-specific:
// confirmed events traversal, cheaters detection.
// Use this structure if need a general-purpose consensus. Instead, use lower-level abft.Orderer.
type Lachesis struct {
	store *Store
	input EventSource

	election *election.Election
	dagIndex DagIndex
	crit     func(error)

	beginBlockFn    lachesis.BeginBlockFn
	applyAtroposFn  ApplyAtroposFn
	epochDBLoadedFn EpochDBLoadedFn
}

// NewLachesis creates Lachesis instance.
func NewLachesis(store *Store, input EventSource, dagIndex DagIndex, crit func(error)) *Lachesis {
	p := &Lachesis{
		store:    store,
		input:    input,
		crit:     crit,
		dagIndex: dagIndex,
	}

	return p
}

func (p *Lachesis) confirmEvents(frame idx.Frame, atropos hash.Event, onEventConfirmed func(dag.Event)) error {
	err := p.dfsSubgraph(atropos, func(e dag.Event) bool {
		decidedFrame := p.store.GetEventConfirmedOn(e.ID())
		if decidedFrame != 0 {
			return false
		}
		// mark all the walked events as confirmed
		p.store.SetEventConfirmedOn(e.ID(), frame)
		if onEventConfirmed != nil {
			onEventConfirmed(e)
		}
		return true
	})
	return err
}

func (p *Lachesis) applyAtropos(decidedFrame idx.Frame, atropos hash.Event) *pos.Validators {
	atroposVecClock := p.dagIndex.GetMergedHighestBefore(atropos)

	validators := p.store.GetValidators()
	// cheaters are ordered deterministically
	cheaters := make([]idx.ValidatorID, 0, validators.Len())
	for creatorIdx, creator := range validators.SortedIDs() {
		if atroposVecClock.Get(idx.Validator(creatorIdx)).IsForkDetected() {
			cheaters = append(cheaters, creator)
		}
	}

	if p.beginBlockFn == nil {
		return nil
	}
	blockCallback := p.beginBlockFn(&lachesis.Block{
		Atropos:  atropos,
		Cheaters: cheaters,
	})

	// traverse newly confirmed events
	err := p.confirmEvents(decidedFrame, atropos, blockCallback.ApplyEvent)
	if err != nil {
		p.crit(err)
	}

	if blockCallback.EndBlock != nil {
		return blockCallback.EndBlock()
	}
	return nil
}

func (p *Lachesis) Bootstrap(callback lachesis.BeginBlockFn) error {
	return p.BootstrapWithOrderer(callback, p.GetApplyAtroposFn(), nil)
}

func (p *Lachesis) BootstrapWithOrderer(beginBlockFn lachesis.BeginBlockFn, applyAtroposFn ApplyAtroposFn, epochDBLoadedFn EpochDBLoadedFn) error {
	if p.election != nil {
		return errors.New("already bootstrapped")
	}
	// block handlers must be set before p.handleElection
	p.applyAtroposFn = applyAtroposFn
	p.epochDBLoadedFn = epochDBLoadedFn
	p.beginBlockFn = beginBlockFn

	// restore current epoch DB
	err := p.loadEpochDB()
	if err != nil {
		return err
	}
	if p.epochDBLoadedFn != nil {
		p.epochDBLoadedFn(p.store.GetEpoch())
	}
	p.election = election.New(p.store.GetValidators(), p.store.GetLastDecidedFrame()+1, p.dagIndex.ForklessCause, p.store.GetFrameRoots)

	// events reprocessing
	_, err = p.bootstrapElection()
	return nil
}

func (p *Lachesis) GetApplyAtroposFn() ApplyAtroposFn {
	return p.applyAtropos
}
