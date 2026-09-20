package genex

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
)

// Sampler visits every generator index once in a seed-dependent order.
// Distinct indices may still produce identical output for overlapping patterns.
type Sampler struct {
	Gen Generator
	// Index is the position in the sampling sequence and may be set to seek.
	Index big.Int
	Seed  int64
}

// Next appends the next sample and advances Index. It returns false without
// changing w or Index when Index is outside [0, Gen.Count()).
func (s *Sampler) Next(w *bytes.Buffer) bool {
	count := s.Gen.Count()
	if s.Index.Sign() < 0 || s.Index.Cmp(count) >= 0 {
		return false
	}

	var value, mask big.Int
	value.Set(&s.Index)
	mask.Sub(count, big.NewInt(1))
	halfBits := uint((mask.BitLen() + 1) / 2)
	if halfBits > 0 {
		mask.Lsh(big.NewInt(1), halfBits)
		mask.Sub(&mask, big.NewInt(1))
		// Cycle walking restricts the permutation to the generator's range.
		for {
			s.permute(&value, halfBits, &mask)
			if value.Cmp(count) < 0 {
				break
			}
		}
	}

	// Generator.Index consumes its argument, so keep the sequence position separate.
	s.Gen.Index(w, &value)
	s.Index.Add(&s.Index, big.NewInt(1))
	return true
}

func (s *Sampler) permute(value *big.Int, halfBits uint, mask *big.Int) {
	var left, right, mixed, block big.Int
	left.Rsh(value, halfBits)
	right.And(value, mask)
	var header [24]byte
	binary.BigEndian.PutUint64(header[:8], uint64(s.Seed))
	for round := uint64(0); round < 8; round++ {
		binary.BigEndian.PutUint64(header[8:16], round)
		rightBytes := right.Bytes()
		mixed.SetInt64(0)
		// Expand the round function for generators with more than 512 index bits.
		for offset := uint(0); offset < halfBits; offset += sha256.Size * 8 {
			binary.BigEndian.PutUint64(header[16:], uint64(offset/(sha256.Size*8)))
			h := sha256.New()
			h.Write(header[:])
			h.Write(rightBytes)
			var digest [sha256.Size]byte
			block.SetBytes(h.Sum(digest[:0]))
			block.Lsh(&block, offset)
			mixed.Or(&mixed, &block)
		}
		mixed.And(&mixed, mask)
		mixed.Xor(&mixed, &left)
		left.Set(&right)
		right.Set(&mixed)
	}
	value.Lsh(&left, halfBits)
	value.Or(value, &right)
}

// NewSampler starts a reproducible sampling sequence at index zero.
func NewSampler(g Generator, seed int64) *Sampler {
	return &Sampler{Gen: g, Seed: seed}
}
