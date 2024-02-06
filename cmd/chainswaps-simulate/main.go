package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/starius/chainswaps"
)

var (
	inBlocks = flag.Int64("in-blocks", 100,
		"Number of blocks to be mined in input blockchain "+
			"(known exactly from incoming invoice)")
	outBlocks = flag.Int64("out-blocks", 42,
		"Max number of blocks to be mined in output blockchain "+
			"(will be set when paying the outgoing invoice)")
	inInterval = flag.Duration("in-interval", 150*time.Second,
		"Average delay between blocks in input blockchain "+
			"(ok to set lower than it is)")
	outInterval = flag.Duration("out-interval", 600*time.Second,
		"Average delay between blocks in output blockchain "+
			"(ok to set higher than it is)")
	inFixed = flag.Bool("in-fixed", false,
		"Input blockchain produces blocks with fixed intervals")
	outFixed = flag.Bool("out-fixed", false,
		"Output blockchain produces blocks with fixed intervals")
	inBlocksReserve = flag.Int64("in-blocks-reserve", 5,
		"Reserve number in input blockchain to let force close tx "+
			"to confirm (ok to set higher)")
	timeReserve = flag.Duration("time-reserve", time.Minute,
		"Leeway to have time to settle incoming invoice after learning"+
			" knowing preimage from outgoing payment (ok to set higher)")
	targetPvalue = flag.Float64("p-value", 1.0/1000000000.0,
		"Target probability of not having enough time to collect "+
			"the money from incoming channel (must be near zero)")
	trials = flag.Int("trials", 1000000,
		"Number of trials in each blockchain simulation "+
			"(higher - more precise)")
	calibration = flag.Bool("calibrate", false,
		"Calibrate max out-blocks to make pvalue <= target pvalue")
	calibrationLeeway = flag.Duration("calibration-leeway", 20*time.Minute,
		"Additional time to add to time reserve during calibration")
)

func main() {
	flag.Parse()

	log.Printf("Incoming blockchain mines a block in %s.", *inInterval)
	log.Printf("Outgoing blockchain mines a block in %s.", *outInterval)

	s := &chainswaps.Swap{
		InBlocks:         *inBlocks,
		OutBlocks:        *outBlocks,
		InInterval:       *inInterval,
		OutInterval:      *outInterval,
		InFixedInterval:  *inFixed,
		OutFixedInterval: *outFixed,
		InBlocksReserve:  *inBlocksReserve,
		TimeReserve:      *timeReserve,
		TargetPvalue:     *targetPvalue,
		Trials:           *trials,
		ExpFloat64:       rand.ExpFloat64,
	}

	if *calibration {
		s.TimeReserve += *calibrationLeeway
		log.Printf("Increased time reserve by %s to %v to calibrate.",
			*calibrationLeeway, s.TimeReserve)
		log.Printf("Calibration for pvalue %v...", s.TargetPvalue)
		s.Calibrate()
		s.TimeReserve -= *calibrationLeeway
		log.Printf("Restored time reserve to %v after calibration.",
			s.TimeReserve)
		log.Printf("The number of blocks in outgoing blockchain: %d.",
			s.OutBlocks)
	}

	log.Printf("Simulation for %d blocks (-%d blocks reserved for "+
		"confirmation) of incoming blockchain, %d blocks of outgoing "+
		"blockchain, time reserve %v...", s.InBlocks, s.InBlocksReserve,
		s.OutBlocks, s.TimeReserve)
	pSim := s.Simulate()
	log.Printf("In simulation the share of being out of time is %v.", pSim)
	if pSim > *targetPvalue {
		log.Fatal("Too high pvalue! BAD!")
	} else {
		log.Printf("GOOD pvalue from simulation!")
	}
}
