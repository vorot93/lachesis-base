package abft

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
)

const dsKey = "d"

// SetLastDecidedState save LastDecidedState.
// LastDecidedState is seldom read; so no cache.
func (s *Store) SetLastDecidedState(v *LastDecidedState) {
	s.set(s.table.LastDecidedState, []byte(dsKey), v)
}

// GetLastDecidedState returns stored LastDecidedState.
// State is seldom read; so no cache.
func (s *Store) GetLastDecidedState() *LastDecidedState {
	w, exists := s.get(s.table.LastDecidedState, []byte(dsKey), &LastDecidedState{}).(*LastDecidedState)
	if !exists {
		s.crit(ErrNoGenesis)
	}

	return w
}

func (s *Store) GetLastDecidedFrame() idx.Frame {
	return s.GetLastDecidedState().LastDecidedFrame
}
