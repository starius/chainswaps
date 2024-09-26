package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/starius/chainswaps"
)

var (
	inBlocksMin     = flag.Int64("in-blocks-min", 30, "Min input blocks")
	inBlocksMax     = flag.Int64("in-blocks-max", 5000, "Max input blocks")
	inBlocksReserve = flag.Int64("in-blocks-reserve", 5,
		"Reserve number in input blockchain to let force close tx "+
			"to confirm (ok to set higher)")
	timeReserve = flag.Duration("time-reserve", time.Minute,
		"Leeway to have time to settle incoming invoice after learning"+
			" knowing preimage from outgoing payment "+
			"(ok to set higher)")
	targetPvalue = flag.Float64("p-value", 1.0/1000000000.0,
		"Target probability of not having enough time to collect "+
			"the money from incoming channel (must be near zero)")
	integrationPoints = flag.Int("integration-points", 1000000,
		"Number of integration points for numerical integration "+
			"(higher - more precise)")
	dev = flag.Float64("interval-deviation", 0.25,
		"Percentage of deviation of interval from canonical value")
	step = flag.Float64("interval-step", 0.01,
		"Percentage step of deviation of interval")
	outDir = flag.String("out-dir", ".",
		"Directory to write output files")
)

func main() {
	flag.Parse()

	swaps := []struct {
		name string
		swap chainswaps.Swap
	}{
		{
			name: "ltc_to_btc",
			swap: chainswaps.Swap{
				InInterval:        150 * time.Second,
				OutInterval:       600 * time.Second,
				InBlocksReserve:   *inBlocksReserve,
				TimeReserve:       *timeReserve,
				TargetPvalue:      *targetPvalue,
				IntegrationPoints: *integrationPoints,
			},
		},
		{
			name: "btc_to_ltc",
			swap: chainswaps.Swap{
				InInterval:        600 * time.Second,
				OutInterval:       150 * time.Second,
				InBlocksReserve:   *inBlocksReserve,
				TimeReserve:       *timeReserve,
				TargetPvalue:      *targetPvalue,
				IntegrationPoints: *integrationPoints,
			},
		},
	}

	for _, c := range swaps {
		log.Printf("Starting swap %s.", c.name)

		table := chainswaps.BuildTable(
			&c.swap,
			int32(*inBlocksMin), int32(*inBlocksMax),
			*dev, *step,
		)

		var buf bytes.Buffer
		if err := table.TSV(&buf); err != nil {
			log.Fatal(err)
		}
		path := filepath.Join(*outDir, c.name+".tsv")
		if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
			log.Fatalf("Failed to write to %s: %v.", path, err)
		}
	}
}
