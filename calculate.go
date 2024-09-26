package chainswaps

import (
	//"fmt"
	"math"
	"sort"
	"time"
)

// Erlang distribution.
type erlangDist struct {
	k       int64
	lambda  float64
	lnConst float64
	kMinus1 float64
}

func newErlangDist(k int64, lambda float64) erlangDist {
	// Precalculate log(lambda^k / (k-1)!).
	lnConst := float64(k) * math.Log(lambda)
	for i := int64(2); i <= k-1; i++ {
		lnConst -= math.Log(float64(i))
	}
	return erlangDist{
		k:       k,
		lambda:  lambda,
		lnConst: lnConst,
		kMinus1: float64(k - 1),
	}
}

// Returns PDF of Erlang distribution.
// lambda^k * x^(k-1) * exp(-lambda * x) / (k-1)!
func (e erlangDist) PDF(x float64) float64 {
	if x == 0 {
		return math.Exp(e.lnConst)
	}
	return math.Exp(e.lnConst + e.kMinus1*math.Log(x) - e.lambda*x)
}

func (e erlangDist) Avg() float64 {
	return float64(e.k) / e.lambda
}

// Returns 1-CDF of Erlang distribution.
// CDF = 1 - exp(-lambda*x) * sum n=0 to (k-1) of 1/n! * (lambda*x)^n.
// To assess the right tail it is more precise than PDF numerical integration.
func (e erlangDist) OneMinusCDF(x float64) float64 {
	lambdaX := e.lambda * x
	lnLambdaX := math.Log(lambdaX)

	// The first item (for n=0) is 1. Note that 0^0=1.
	lnItem := float64(0)
	sum := math.Exp(-lambdaX)
	for n := int64(1); n <= e.k-1; n++ {
		lnItem += lnLambdaX
		lnItem -= math.Log(float64(n))
		sum += math.Exp(-lambdaX + lnItem)

		//fmt.Printf("%0.400f\n", lambdaX)
		//fmt.Printf("%0.400f\n", item)
		//fmt.Println(item, sum, lambdaX, float64(n))
	}

	if sum > 1 {
		sum = 1
	}

	return sum
}

func (s *Swap) Calculate() float64 {
	inInterval := float64(s.InInterval) / float64(time.Second)
	outInterval := float64(s.OutInterval) / float64(time.Second)

	inLambda := 1 / inInterval
	outLambda := 1 / outInterval
	timeReserve := float64(s.TimeReserve) / float64(time.Second)
	inBlocks := s.InBlocks - s.InBlocksReserve

	//maxT := outInterval * float64(s.OutBlocks) * 20
	//step := maxT / float64(s.IntegrationPoints)

	inErlang := newErlangDist(inBlocks, inLambda)
	outErlang := newErlangDist(s.OutBlocks, outLambda)

	maxT := findMaxT(inErlang, outErlang)
	background := 1 / maxT

	// Calculate CDF(in) at timeReserve.
	inCDF := 1 - inErlang.OneMinusCDF(timeReserve)
	//fmt.Printf("inBlocks=%v inLambda=%v timeReserve=%v inCDF=%v\n", inBlocks, inLambda, timeReserve, inCDF)
	//for t := float64(0); t < timeReserve; t += step {
	//	width := step
	//	if t+step > timeReserve {
	//		// Cut last item.
	//		width = timeReserve - t
	//	}
	//	inCDF += inErlang.PDF(t) * width
	//}

	var pvalue float64

	//ttt := outErlang.OneMinusCDF(424455)
	//fmt.Println(ttt)

	//panic(1)

	//outCDF := float64(0)

	// 3 is sum(pdf(in)) + sum(pdf(out)) + sum(background).
	stepNom := 3 / float64(s.IntegrationPoints)

	prevOutPDF := outErlang.PDF(0)
	prevInPDF := inErlang.PDF(timeReserve)

	var t float64
	for i := 0; i < s.IntegrationPoints; i++ {
		//for t := float64(0); t < maxT; t += step {

		// The higher PDFs, the smaller the step.
		// Background is needed to avoid division by 0
		// and super-huge jumps. The value of background is
		// such that in the end we achieve approximately maxT.
		step := stepNom / (background + prevOutPDF + prevInPDF)
		nextT := t + step

		outPDF := outErlang.PDF(nextT)
		inPDF := inErlang.PDF(nextT + timeReserve)

		// Use triangular approximation.
		//outP := (prevOutPDF+outPDF) / 2 * step
		inP := (prevInPDF + inPDF) / 2 * step

		//fmt.Println(t, outPDF, outP, outCDF, 1-outErlang.OneMinusCDF(t), inPDF, inP, outPDF+inPDF, step)
		//fmt.Println(outErlang.PDF(t), (1-outCDF), pvalue/100)

		square := trapezeConvolutionIgregral(step, prevOutPDF, outPDF-prevOutPDF, inCDF, inPDF)
		pvalue += square

		//fmt.Printf("t=%v step=%v prevOutPDF=%v outPDF-prevOutPDF=%v inCDF=%v inPDF=%v square=%v pvalue=%v\n", t, step, prevOutPDF, outPDF-prevOutPDF, inCDF, inPDF, square, pvalue)

		//outCDF += outP
		inCDF += inP

		prevOutPDF = outPDF
		prevInPDF = inPDF
		t = nextT
	}

	// Add leftover of out CDF, because we are interested in upper bound.
	leftover := outErlang.OneMinusCDF(t)
	//fmt.Println("leftover", leftover)
	pvalue += leftover

	return pvalue
}

// Find T for which CDF of both in and out is 0 in float64.
func findMaxT(in, out erlangDist) float64 {
	// On average, out < in, so check in first to save computation.
	t := in.Avg() * 10
	for in.OneMinusCDF(t) > 0 {
		t *= 1.5
	}
	for out.OneMinusCDF(t) > 0 {
		t *= 1.5
	}
	return t
}

func trapezeConvolutionIgregral(step, xPDF, xPDFdelta, yCDF, yCDFdelta float64) float64 {
	step2 := step * step
	step3 := step2 * step
	return step*xPDF*yCDF +
		step2/2*(xPDF*yCDFdelta+yCDF*xPDFdelta) +
		step3/3*xPDFdelta*yCDFdelta
}

// Calibrate sets OutBlocks according to TargetPvalue and other parameters.
func (s *Swap) CalibrateWithCalculation() {
	inBlocks := s.InBlocks - s.InBlocksReserve

	max := inBlocks * int64(s.InInterval) / int64(s.OutInterval)
	max += 1 // Round up.

	s.OutBlocks = int64(sort.Search(int(max), func(outBlocks int) bool {
		s.OutBlocks = int64(outBlocks)
		return s.Calculate() >= s.TargetPvalue
	})) - 1
}
