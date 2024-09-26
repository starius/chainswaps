package chainswaps

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Row struct {
	InInterval time.Duration
	OutBlocks  []int32
}

type Table struct {
	Header []int32
	Rows   []Row
}

func BuildTable(swap *Swap, inBlocksMin, inBlocksMax int32,
	dev, step float64) Table {

	inInterval0 := swap.InInterval
	width := inBlocksMax - inBlocksMin + 1

	header := make([]int32, 0, width)
	for i := inBlocksMin; i <= inBlocksMax; i++ {
		header = append(header, i)
	}

	table := Table{
		Header: header,
	}
	for d := 1 - dev; d <= 1+dev; d += step {
		row := Row{
			InInterval: time.Duration(float64(inInterval0) * d),
		}
		table.Rows = append(table.Rows, row)
	}

	totalWork := len(table.Rows) * int(width)
	var itemsDone atomic.Int64
	incrementProgress := func(items int) {
		done := itemsDone.Add(int64(items))
		//if done&1023 == 0 || done == int64(totalWork) {
		fmt.Fprintf(os.Stderr, "Done: %d/%d = %.2f%%\r",
			done, totalWork,
			float64(done)/float64(totalWork)*100)
		//}
	}

	var wg sync.WaitGroup
	wg.Add(len(table.Rows))
	for j := 0; j < len(table.Rows); j++ {
		go func(j int) {
			defer wg.Done()

			swap1 := *swap
			swap1.InInterval = table.Rows[j].InInterval
			cells := make([]int32, 0, width)

			// Calibrate for the first cell.
			swap1.InBlocks = int64(inBlocksMin)
			swap1.CalibrateWithCalculation()
			if swap1.OutBlocks < 0 {
				swap1.OutBlocks = 0
			}
			cells = append(cells, int32(swap1.OutBlocks))
			incrementProgress(1)

			// Calculate next cells using previous value.

			for i := inBlocksMin + 1; i <= inBlocksMax; i++ {
				swap1.InBlocks = int64(i)
				for swap1.Calculate() < swap1.TargetPvalue {
					swap1.OutBlocks++
				}
				swap1.OutBlocks--
				if swap1.OutBlocks < 0 {
					swap1.OutBlocks = 0
				}
				cells = append(cells, int32(swap1.OutBlocks))
				incrementProgress(1)
			}
			table.Rows[j].OutBlocks = cells
		}(j)
	}

	wg.Wait()

	fmt.Fprintf(os.Stderr, "\n")

	return table
}

func (t *Table) TSV(w io.Writer) error {
	tsv := csv.NewWriter(w)
	tsv.Comma = '\t'

	header := make([]string, 0, len(t.Header)+1)
	header = append(header, "")
	for _, i := range t.Header {
		header = append(header, strconv.Itoa(int(i)))
	}
	if err := tsv.Write(header); err != nil {
		return err
	}

	for _, r := range t.Rows {
		row := make([]string, 0, len(r.OutBlocks)+1)
		row = append(row, r.InInterval.String())
		for _, c := range r.OutBlocks {
			row = append(row, strconv.Itoa(int(c)))
		}
		if err := tsv.Write(row); err != nil {
			return err
		}
	}

	tsv.Flush()

	if err := tsv.Error(); err != nil {
		return err
	}

	return nil
}
