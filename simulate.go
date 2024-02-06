package chainswaps

import (
	"sort"
)

func (s *Swap) randErlang(r int64) float64 {
	var sum float64
	for i := int64(0); i < r; i++ {
		sum += s.ExpFloat64()
	}
	return sum
}

func (s *Swap) randErlangSeries(r int64) []float64 {
	var sum float64
	res := make([]float64, r)
	for i := int64(0); i < r; i++ {
		res[i] = sum
		sum += s.ExpFloat64()
	}
	return res
}

func (s *Swap) simulateTime(blocks int64, fixed bool) float64 {
	if fixed {
		return float64(blocks)
	} else {
		return s.randErlang(blocks)
	}
}

type sample struct {
	t  float64
	in bool
}

// Calculate probility of in time < out time given in and out samples.
func probInLessThenOut(samples []sample) float64 {
	sort.Slice(samples, func(i, j int) bool {
		return samples[i].t < samples[j].t
	})

	nom := 0
	seenIns := 0

	for _, s := range samples {
		if s.in {
			seenIns++
		} else {
			nom += seenIns
		}
	}

	seenOuts := len(samples) - seenIns

	return float64(nom) / float64(seenIns*seenOuts)
}

// Simulate runs Trials simulations for each blockchain and returns
// the share of combinations where incoming expired before outgoing,
// taking into account InBlocksReserve and TimeReserve.
func (s *Swap) Simulate() float64 {
	// Assign interval of incoming to 1.
	scale := float64(s.OutInterval) / float64(s.InInterval)
	timeReserve := float64(s.TimeReserve) / float64(s.InInterval)

	samples := make([]sample, 0, 2*s.Trials)

	inBlocks := s.InBlocks - s.InBlocksReserve

	for i := 0; i < s.Trials; i++ {
		in := s.simulateTime(inBlocks, s.InFixedInterval) - timeReserve
		out := s.simulateTime(s.OutBlocks, s.OutFixedInterval) * scale
		samples = append(samples,
			sample{t: in, in: true},
			sample{t: out},
		)
	}

	return probInLessThenOut(samples)
}

// Calibrate sets OutBlocks according to TargetPvalue and other parameters.
func (s *Swap) Calibrate() {
	// Assign interval of incoming to 1.
	scale := float64(s.OutInterval) / float64(s.InInterval)
	timeReserve := float64(s.TimeReserve) / float64(s.InInterval)

	max := s.InBlocks * int64(s.InInterval) / int64(s.OutInterval)
	max += 1 // Round up.

	inBlocks := s.InBlocks - s.InBlocksReserve

	inSamples := make([]float64, s.Trials)
	outSeries := make([][]float64, s.Trials)
	for i := 0; i < s.Trials; i++ {
		if !s.InFixedInterval {
			in := s.randErlang(inBlocks) - timeReserve
			inSamples[i] = in
		}
		if !s.OutFixedInterval {
			outSeries[i] = s.randErlangSeries(max)
		}
	}

	s.OutBlocks = int64(sort.Search(int(max), func(outBlocks int) bool {
		samples := make([]sample, 0, 2*s.Trials)

		for i := 0; i < s.Trials; i++ {
			var in, out float64
			if s.InFixedInterval {
				in = float64(inBlocks) - timeReserve
			} else {
				in = inSamples[i]
			}
			if s.OutFixedInterval {
				out = float64(outBlocks) * scale
			} else {
				out = outSeries[i][outBlocks] * scale
			}
			samples = append(samples,
				sample{t: in, in: true},
				sample{t: out},
			)
		}

		return probInLessThenOut(samples) >= s.TargetPvalue
	})) - 1
}
