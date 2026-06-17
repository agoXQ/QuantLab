package indicators

import (
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// KDJ and BOLL share their lookback logic with rolling helpers above.
//
// KDJ requires the high/low/close triplet but the Formula Engine MVP works on
// a single price series at a time. We approximate the n-period high/low using
// the same series passed in (which is typically CLOSE). When the evaluator is
// wired to OHLCV bars it will pass a Bar-aware indicator instead.

func kdjK(args []domainInd.Arg) (series.Series, error) {
	k, _, _, err := kdjLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return k, nil
}

func kdjD(args []domainInd.Arg) (series.Series, error) {
	_, d, _, err := kdjLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return d, nil
}

func kdjJ(args []domainInd.Arg) (series.Series, error) {
	_, _, j, err := kdjLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return j, nil
}

func kdjLines(args []domainInd.Arg) (series.Series, series.Series, series.Series, error) {
	if err := requireArgs("KDJ", len(args), 4); err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("KDJ", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	n, err := asInt("KDJ", 1, args[1])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	m1, err := asInt("KDJ", 2, args[2])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	m2, err := asInt("KDJ", 3, args[3])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}

	hh := rollingExtreme(s, n, true)
	ll := rollingExtreme(s, n, false)
	rsv := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		hi, lo, c := hh.At(i), ll.At(i), s.At(i)
		if math.IsNaN(hi) || math.IsNaN(lo) || math.IsNaN(c) || hi == lo {
			rsv[i] = math.NaN()
			continue
		}
		rsv[i] = (c - lo) / (hi - lo) * 100
	}
	rsvSeries, _ := series.NewSeries(s.Timestamps(), rsv)
	k := smaSeries(rsvSeries, m1, 1)
	d := smaSeries(k, m2, 1)
	j := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if math.IsNaN(k.At(i)) || math.IsNaN(d.At(i)) {
			j[i] = math.NaN()
			continue
		}
		j[i] = 3*k.At(i) - 2*d.At(i)
	}
	jSeries, _ := series.NewSeries(s.Timestamps(), j)
	return k, d, jSeries, nil
}

// bollMid implements BOLL_MID(series, n, k) returning the middle band (MA).
func bollMid(args []domainInd.Arg) (series.Series, error) {
	mid, _, _, err := bollLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return mid, nil
}

// bollUp implements BOLL_UP(series, n, k) returning the upper band.
func bollUp(args []domainInd.Arg) (series.Series, error) {
	_, up, _, err := bollLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return up, nil
}

// bollDown implements BOLL_DOWN(series, n, k) returning the lower band.
func bollDown(args []domainInd.Arg) (series.Series, error) {
	_, _, lo, err := bollLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return lo, nil
}

func bollLines(args []domainInd.Arg) (series.Series, series.Series, series.Series, error) {
	if err := requireArgs("BOLL", len(args), 3); err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("BOLL", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	n, err := asInt("BOLL", 1, args[1])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	k, err := asFloat("BOLL", 2, args[2])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}

	mid := rollingMean(s, n)
	stdv := rollingStd(s, n)
	up := make([]float64, s.Len())
	lo := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		m := mid.At(i)
		sd := stdv.At(i)
		if math.IsNaN(m) || math.IsNaN(sd) {
			up[i] = math.NaN()
			lo[i] = math.NaN()
			continue
		}
		up[i] = m + k*sd
		lo[i] = m - k*sd
	}
	upSeries, _ := series.NewSeries(s.Timestamps(), up)
	loSeries, _ := series.NewSeries(s.Timestamps(), lo)
	return mid, upSeries, loSeries, nil
}
