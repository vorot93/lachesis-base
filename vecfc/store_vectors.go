package vecfc

import (
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
)

func (vi *Index) getBytes(table kvdb.Store, id hash.Event) []byte {
	key := id.Bytes()
	b, err := table.Get(key)
	if err != nil {
		vi.crit(err)
	}
	return b
}

func (vi *Index) setBytes(table kvdb.Store, id hash.Event, b []byte) {
	key := id.Bytes()
	err := table.Put(key, b)
	if err != nil {
		vi.crit(err)
	}
}

// GetLowestAfter reads the vector from DB
func (vi *Index) GetLowestAfter(id hash.Event) *LowestAfterSeq {
	b := LowestAfterSeq(vi.getBytes(vi.table.LowestAfterSeq, id))
	if b == nil {
		return nil
	}
	return &b
}

// GetHighestBefore reads the vector from DB
func (vi *Index) GetHighestBefore(id hash.Event) *HighestBeforeSeq {
	b := HighestBeforeSeq(vi.getBytes(vi.table.HighestBeforeSeq, id))
	if b == nil {
		return nil
	}
	return &b
}

// SetLowestAfter stores the vector into DB
func (vi *Index) SetLowestAfter(id hash.Event, seq *LowestAfterSeq) {
	vi.setBytes(vi.table.LowestAfterSeq, id, *seq)
}

// SetHighestBefore stores the vectors into DB
func (vi *Index) SetHighestBefore(id hash.Event, seq *HighestBeforeSeq) {
	vi.setBytes(vi.table.HighestBeforeSeq, id, *seq)
}
