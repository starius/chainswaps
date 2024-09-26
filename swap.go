package chainswaps

import (
	"time"
)

type Swap struct {
	// Number of blocks to be mined in input blockchain.
	// Its exact value is known from incoming invoice.
	InBlocks int64

	// Max number of blocks to be mined in output blockchain.
	// It is produced in calibration or passed from input.
	// It will be set to --cltv_limit when paying the outgoing invoice.
	OutBlocks int64 // Produced in calibration.

	// Average delay between blocks in input blockchain.
	// It is ok to set lower than it is.
	InInterval time.Duration

	// Average delay between blocks in output blockchain.
	// It is ok to set higher than it is.
	OutInterval time.Duration

	// If input blockchain has fixed block intervals.
	InFixedInterval bool

	// If output blockchain has fixed block intervals.
	OutFixedInterval bool

	// Reserve number in input blockchain to let force close tx to confirm.
	// It is ok to set higher.
	InBlocksReserve int64

	// Leeway to have time to settle incoming invoice after learning knowing
	// preimage from outgoing payment. It is ok to set higher.
	TimeReserve time.Duration

	// Target probability of not having enough time to collect the money
	// from incoming channel used during calibration. Must be near zero.
	TargetPvalue float64

	// Number of trials in each blockchain simulation.
	// Higher is more precise.
	Trials int

	// ExpFloat64 returns an exponentially distributed float64 in the range
	// (0, +math.MaxFloat64] with an exponential distribution whose rate
	// parameter (lambda) is 1 and whose mean is 1/lambda (1).
	ExpFloat64 func() float64

	// Number of integration points for numerical integration.
	// Used by Calculate() method. Higher is more precise.
	IntegrationPoints int
}
