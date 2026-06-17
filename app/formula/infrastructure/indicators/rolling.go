package indicators

import (
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// ma implements MA(series, period): simple moving average.
func ma(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("MA", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("MA", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("MA", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rollingMean(s, n), nil
}

// avg is an alias for MA. Kept as a separate registration so future versions
// can diverge if the DSL semantics change.
func avg(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("AVG", len(args), 2); err != nil {
		return series.Series{}, err
	}
	return ma(args)
}

// sum implements SUM(series, n): rolling sum.
func sum(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("SUM", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("SUM", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("SUM", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rollingSum(s, n), nil
}

// std implements STD(series, n): rolling sample standard deviation.
func std(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("STD", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("STD", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("STD", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rollingStd(s, n), nil
}

// hhv implements HHV(series, n): highest value over a rolling window.
func hhv(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("HHV", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("HHV", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("HHV", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rollingExtreme(s, n, true), nil
}

// llv implements LLV(series, n): lowest value over a rolling window.
func llv(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("LLV", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("LLV", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("LLV", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rollingExtreme(s, n, false), nil
}

// count implements COUNT(condition, n): number of true samples in the
// trailing window of length n.
func count(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("COUNT", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	cond, err := asSeries("COUNT", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("COUNT", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}

	out := make([]float64, cond.Len())
	for i := 0; i < cond.Len(); i++ {
		if i+1 < n {
			out[i] = math.NaN()
			continue
		}
		c := 0
		for j := i - n + 1; j <= i; j++ {
			v := cond.At(j)
			if !math.IsNaN(v) && v != 0 {
				c++
			}
		}
		out[i] = float64(c)
	}
	return series.NewSeries(cond.Timestamps(), out)
}

// --- shared helpers ---

func rollingMean(s series.Series, n int) series.Series {
	out := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i+1 < n {
			out[i] = math.NaN()
			continue
		}
		var acc float64
		var k int
		valid := true
		for j := i - n + 1; j <= i; j++ {
			v := s.At(j)
			if math.IsNaN(v) {
				valid = false
				break
			}
			acc += v
			k++
		}
		if !valid {
			out[i] = math.NaN()
			continue
		}
		out[i] = acc / float64(k)
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func rollingSum(s series.Series, n int) series.Series {
	out := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i+1 < n {
			out[i] = math.NaN()
			continue
		}
		var acc float64
		valid := true
		for j := i - n + 1; j <= i; j++ {
			v := s.At(j)
			if math.IsNaN(v) {
				valid = false
				break
			}
			acc += v
		}
		if !valid {
			out[i] = math.NaN()
			continue
		}
		out[i] = acc
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func rollingStd(s series.Series, n int) series.Series {
	out := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i+1 < n {
			out[i] = math.NaN()
			continue
		}
		var sum, sumSq float64
		valid := true
		for j := i - n + 1; j <= i; j++ {
			v := s.At(j)
			if math.IsNaN(v) {
				valid = false
				break
			}
			sum += v
			sumSq += v * v
		}
		if !valid {
			out[i] = math.NaN()
			continue
		}
		mean := sum / float64(n)
		variance := (sumSq - float64(n)*mean*mean) / float64(n-1)
		if variance < 0 {
			variance = 0
		}
		out[i] = math.Sqrt(variance)
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func rollingExtreme(s series.Series, n int, max bool) series.Series {
	out := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i+1 < n {
			out[i] = math.NaN()
			continue
		}
		ext := math.NaN()
		for j := i - n + 1; j <= i; j++ {
			v := s.At(j)
			if math.IsNaN(v) {
				ext = math.NaN()
				break
			}
			if math.IsNaN(ext) {
				ext = v
				continue
			}
			if max && v > ext {
				ext = v
			}
			if !max && v < ext {
				ext = v
			}
		}
		out[i] = ext
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}
