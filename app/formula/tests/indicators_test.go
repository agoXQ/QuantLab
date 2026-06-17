package tests

import (
	"math"
	"testing"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
	infraInd "github.com/agoXQ/QuantLab/app/formula/infrastructure/indicators"
)

const eps = 1e-9

func indicatorLib(t *testing.T, name string) domainInd.Indicator {
	t.Helper()
	lib := infraInd.NewLibrary()
	fn, ok := lib.Get(name)
	if !ok {
		t.Fatalf("indicator %s not registered", name)
	}
	return fn
}

func almostEqual(a, b float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= eps
}

func TestIndicators_MA(t *testing.T) {
	fn := indicatorLib(t, "MA")
	in := series.FromValues([]float64{1, 2, 3, 4, 5})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(3)})
	if err != nil {
		t.Fatalf("MA err: %v", err)
	}
	want := []float64{math.NaN(), math.NaN(), 2, 3, 4}
	for i, v := range out.Values() {
		if !almostEqual(v, want[i]) {
			t.Errorf("MA[%d] = %v, want %v", i, v, want[i])
		}
	}
}

func TestIndicators_EMA(t *testing.T) {
	fn := indicatorLib(t, "EMA")
	in := series.FromValues([]float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(3)})
	if err != nil {
		t.Fatalf("EMA err: %v", err)
	}
	// First two are NaN; index 2 is the seed mean of 1+2+3 = 2.
	if !almostEqual(out.At(2), 2) {
		t.Errorf("EMA seed = %v, want 2", out.At(2))
	}
	// EMA[3] = 0.5*4 + 0.5*2 = 3
	if !almostEqual(out.At(3), 3) {
		t.Errorf("EMA[3] = %v, want 3", out.At(3))
	}
}

func TestIndicators_RSI_AllUp(t *testing.T) {
	fn := indicatorLib(t, "RSI")
	in := series.FromValues([]float64{1, 2, 3, 4, 5, 6, 7, 8})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(5)})
	if err != nil {
		t.Fatalf("RSI err: %v", err)
	}
	// All gains -> RSI = 100 from index 5 onwards.
	if !almostEqual(out.At(5), 100) {
		t.Errorf("RSI all-up = %v, want 100", out.At(5))
	}
}

func TestIndicators_HHV_LLV(t *testing.T) {
	hhv := indicatorLib(t, "HHV")
	llv := indicatorLib(t, "LLV")
	in := series.FromValues([]float64{3, 1, 4, 1, 5, 9, 2, 6})
	high, _ := hhv([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(3)})
	low, _ := llv([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(3)})
	if !almostEqual(high.At(5), 9) {
		t.Errorf("HHV[5] = %v, want 9", high.At(5))
	}
	if !almostEqual(low.At(5), 1) {
		t.Errorf("LLV[5] = %v, want 1", low.At(5))
	}
}

func TestIndicators_REF(t *testing.T) {
	fn := indicatorLib(t, "REF")
	in := series.FromValues([]float64{10, 20, 30, 40, 50})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(2)})
	if err != nil {
		t.Fatalf("REF err: %v", err)
	}
	want := []float64{math.NaN(), math.NaN(), 10, 20, 30}
	for i, v := range out.Values() {
		if !almostEqual(v, want[i]) {
			t.Errorf("REF[%d] = %v, want %v", i, v, want[i])
		}
	}
}

func TestIndicators_CROSS(t *testing.T) {
	fn := indicatorLib(t, "CROSS")
	a := series.FromValues([]float64{1, 2, 3, 4, 5})
	b := series.FromValues([]float64{5, 4, 3, 2, 1})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(a), domainInd.SeriesArg(b)})
	if err != nil {
		t.Fatalf("CROSS err: %v", err)
	}
	// First bar is NaN; prev a<=b at i=2 (3<=3), then a>b at i=3 -> 1.
	if !almostEqual(out.At(3), 1) {
		t.Errorf("CROSS[3] = %v, want 1", out.At(3))
	}
	if !almostEqual(out.At(4), 0) {
		t.Errorf("CROSS[4] = %v, want 0", out.At(4))
	}
}

func TestIndicators_BollMid(t *testing.T) {
	fn := indicatorLib(t, "BOLL_MID")
	in := series.FromValues([]float64{1, 2, 3, 4, 5})
	out, err := fn([]domainInd.Arg{domainInd.SeriesArg(in), domainInd.NumberArg(3), domainInd.NumberArg(2)})
	if err != nil {
		t.Fatalf("BOLL_MID err: %v", err)
	}
	if !almostEqual(out.At(4), 4) {
		t.Errorf("BOLL_MID[4] = %v, want 4", out.At(4))
	}
}

func TestIndicators_ArgCountError(t *testing.T) {
	fn := indicatorLib(t, "MA")
	_, err := fn([]domainInd.Arg{domainInd.SeriesArg(series.FromValues([]float64{1, 2}))})
	if err == nil {
		t.Fatal("expected error for missing period arg")
	}
}

func TestIndicators_PeriodMustBePositive(t *testing.T) {
	fn := indicatorLib(t, "MA")
	_, err := fn([]domainInd.Arg{domainInd.SeriesArg(series.FromValues([]float64{1, 2})), domainInd.NumberArg(-1)})
	if err == nil {
		t.Fatal("expected error for negative period")
	}
}
