package abft

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
)

const (
	FirstFrame = idx.Frame(1)
	FirstEpoch = idx.Epoch(1)
)

// LastDecidedState is for persistent storing.
type LastDecidedState struct {
	// fields can change only after a frame is decided
	LastDecidedFrame idx.Frame
}

type EpochState struct {
	// stored values
	// these values change only after a change of epoch
	Epoch      idx.Epoch
	Validators *pos.Validators
}

func (es EpochState) String() string {
	return fmt.Sprintf("%d/%s", es.Epoch, es.Validators.String())
}

// Reset switches epoch state to a new empty epoch.
func (p *Lachesis) Reset(epoch idx.Epoch, validators *pos.Validators) error {
	p.store.applyGenesis(epoch, validators)
	// reset internal epoch DB
	err := p.resetEpochStore(epoch)
	if err != nil {
		return err
	}
	p.election.Reset(validators, FirstFrame)
	return nil
}

func (p *Lachesis) loadEpochDB() error {
	return p.store.openEpochDB(p.store.GetEpoch())
}
