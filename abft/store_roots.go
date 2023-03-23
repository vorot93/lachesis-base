package abft

import (
	"bytes"
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/abft/election"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

func rootRecordKey(r *election.RootAndSlot) []byte {
	key := bytes.Buffer{}
	key.Write(r.Slot.Frame.Bytes())
	key.Write(r.Slot.Validator.Bytes())
	key.Write(r.ID.Bytes())
	return key.Bytes()
}

// AddRoot stores the new root
// Not safe for concurrent use due to the complex mutable cache!
func (s *Store) AddRoot(selfParentFrame idx.Frame, root dag.Event) {
	for f := selfParentFrame + 1; f <= root.Frame(); f++ {
		s.addRoot(root, f)
	}
}

func (s *Store) addRoot(root dag.Event, frame idx.Frame) {
	r := election.RootAndSlot{
		Slot: election.Slot{
			Frame:     frame,
			Validator: root.Creator(),
		},
		ID: root.ID(),
	}

	if err := s.epochTable.Roots.Put(rootRecordKey(&r), []byte{}); err != nil {
		s.crit(err)
	}
}

const (
	frameSize       = 4
	validatorIDSize = 4
	eventIDSize     = 32
)

// GetFrameRoots returns all the roots in the specified frame
// Not safe for concurrent use due to the complex mutable cache!
func (s *Store) GetFrameRoots(f idx.Frame) []election.RootAndSlot {
	rr := make([]election.RootAndSlot, 0, 100)

	it := s.epochTable.Roots.NewIterator(f.Bytes(), nil)
	defer it.Release()
	for it.Next() {
		key := it.Key()
		if len(key) != frameSize+validatorIDSize+eventIDSize {
			s.crit(fmt.Errorf("roots table: incorrect key len=%d", len(key)))
		}
		r := election.RootAndSlot{
			Slot: election.Slot{
				Frame:     idx.BytesToFrame(key[:frameSize]),
				Validator: idx.BytesToValidatorID(key[frameSize : frameSize+validatorIDSize]),
			},
			ID: hash.BytesToEvent(key[frameSize+validatorIDSize:]),
		}
		if r.Slot.Frame != f {
			s.crit(fmt.Errorf("roots table: invalid frame=%d, expected=%d", r.Slot.Frame, f))
		}

		rr = append(rr, r)
	}
	if it.Error() != nil {
		s.crit(it.Error())
	}

	return rr
}
