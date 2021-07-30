// +build gofuzzbeta

package txpool

import (
	"testing"
)

// https://blog.golang.org/fuzz-beta
// golang.org/s/draft-fuzzing-design
//gotip doc testing
//gotip doc testing.F
//gotip doc testing.F.Add
//gotip doc testing.F.Fuzz

// gotip test -trimpath -v -fuzz=Fuzz -fuzztime=10s ./txpool

func FuzzPromoteStep(f *testing.F) {
	f.Add([]uint8{0b11111, 0b10001, 0b10101, 0b00001, 0b00000}, []uint8{0b11111, 0b10001, 0b10101, 0b00001, 0b00000}, []uint8{0b11111, 0b10001, 0b10101, 0b00001, 0b00000})
	f.Add([]uint8{0b11111}, []uint8{0b11111}, []uint8{0b11110, 0b0, 0b1010})
	f.Fuzz(func(t *testing.T, s1, s2, s3 []uint8) {
		t.Parallel()
		pending := NewSubPool()
		for i := range s1 {
			s1[i] &= 0b11111
			pending.Add(&MetaTx{SubPool: SubPoolMarker(s1[i])})
		}
		baseFee := NewSubPool()
		for i := range s2 {
			s2[i] &= 0b11111
			baseFee.Add(&MetaTx{SubPool: SubPoolMarker(s2[i])})
		}
		queue := NewSubPool()
		for i := range s3 {
			s3[i] &= 0b11111
			queue.Add(&MetaTx{SubPool: SubPoolMarker(s3[i])})
		}
		PromoteStep(pending, baseFee, queue)

		best, worst := pending.Best(), pending.Worst()
		_ = best
		if worst != nil && worst.SubPool < 0b01111 {
			t.Fatalf("Pending worst too small %b, input: %b,%b,%b", worst.SubPool, s1, s2, s3)
		}

		best, worst = baseFee.Best(), baseFee.Worst()
		_ = best
		if worst != nil && worst.SubPool < 0b01111 {
			t.Fatalf("Pending worst too small %b, input: %b,%b,%b", worst.SubPool, s1, s2, s3)
		}

		best, worst = queue.Best(), queue.Worst()
		_ = best
		if worst != nil && worst.SubPool < 0b01111 {
			t.Fatalf("Pending worst too small %b, input: %b,%b,%b", worst.SubPool, s1, s2, s3)
		}

	})
}
