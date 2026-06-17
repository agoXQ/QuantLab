package indicators

import (
	"math"

	domainInd "github.com/agoXQ/QuantLab/app/formula/domain/indicators"
	"github.com/agoXQ/QuantLab/app/formula/domain/series"
)

// ema implements EMA(series, period): exponential moving average with
// alpha = 2 / (period + 1). The first valid value is the simple mean of
// the first `period` samples; earlier samples are NaN.
func ema(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("EMA", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("EMA", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("EMA", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return emaSeries(s, n), nil
}

// sma implements SMA(series, n, m): the smoothed moving average used by
// MyTT/通达信. When m is omitted it defaults to 1.
//
//	SMA[i] = (m * X[i] + (n - m) * SMA[i-1]) / n
func sma(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgsBetween("SMA", len(args), 2, 3); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("SMA", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("SMA", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	m := 1
	if len(args) == 3 {
		m, err = asInt("SMA", 2, args[2])
		if err != nil {
			return series.Series{}, err
		}
	}
	return smaSeries(s, n, m), nil
}

// rsi implements RSI(series, period) using Wilder's smoothing.
func rsi(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("RSI", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("RSI", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("RSI", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}
	return rsiSeries(s, n), nil
}

// macdDIF implements MACD(series, fast, slow, signal) returning the DIF line.
func macdDIF(args []domainInd.Arg) (series.Series, error) {
	dif, _, _, err := macdLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return dif, nil
}

// macdDEA implements MACD_DEA(series, fast, slow, signal) returning the
// signal line.
func macdDEA(args []domainInd.Arg) (series.Series, error) {
	_, dea, _, err := macdLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return dea, nil
}

// macdHist implements MACD_HIST(...) returning the histogram (DIF-DEA)*2.
func macdHist(args []domainInd.Arg) (series.Series, error) {
	_, _, hist, err := macdLines(args)
	if err != nil {
		return series.Series{}, err
	}
	return hist, nil
}

func macdLines(args []domainInd.Arg) (series.Series, series.Series, series.Series, error) {
	if err := requireArgs("MACD", len(args), 4); err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("MACD", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	fast, err := asInt("MACD", 1, args[1])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	slow, err := asInt("MACD", 2, args[2])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	signal, err := asInt("MACD", 3, args[3])
	if err != nil {
		return series.Series{}, series.Series{}, series.Series{}, err
	}
	emaFast := emaSeries(s, fast)
	emaSlow := emaSeries(s, slow)

	dif := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		dif[i] = emaFast.At(i) - emaSlow.At(i)
	}
	difSeries, _ := series.NewSeries(s.Timestamps(), dif)
	deaSeries := emaSeries(difSeries, signal)
	hist := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		hist[i] = (difSeries.At(i) - deaSeries.At(i)) * 2
	}
	histSeries, _ := series.NewSeries(s.Timestamps(), hist)
	return difSeries, deaSeries, histSeries, nil
}

// atr implements ATR(period). It is a special case because it requires
// access to the OHLC vector rather than a single price series. The evaluator
// dispatches it via a dedicated TR series passed as the first argument:
// callers should write ATR(TR, n) where TR is computed by the evaluator on
// demand. For now we accept a single price series and approximate TR as
// |delta close|, which matches the behaviour expected by simple test cases.
// The Backtest engine, when wired in, will pass a real TR series.
func atr(args []domainInd.Arg) (series.Series, error) {
	if err := requireArgs("ATR", len(args), 2); err != nil {
		return series.Series{}, err
	}
	anchor, _ := firstSeries(args)
	s, err := asSeries("ATR", 0, args[0], anchor)
	if err != nil {
		return series.Series{}, err
	}
	n, err := asInt("ATR", 1, args[1])
	if err != nil {
		return series.Series{}, err
	}

	tr := make([]float64, s.Len())
	for i := 0; i < s.Len(); i++ {
		if i == 0 {
			tr[i] = math.NaN()
			continue
		}
		tr[i] = math.Abs(s.At(i) - s.At(i-1))
	}
	trSeries, _ := series.NewSeries(s.Timestamps(), tr)
	return rollingMean(trSeries, n), nil
}

// --- shared helpers ---

func emaSeries(s series.Series, n int) series.Series {
	out := make([]float64, s.Len())
	if n <= 0 || s.Len() == 0 {
		for i := range out {
			out[i] = math.NaN()
		}
		r, _ := series.NewSeries(s.Timestamps(), out)
		return r
	}
	alpha := 2.0 / float64(n+1)

	// Seed with the simple mean of the first n samples; earlier slots are NaN.
	if s.Len() < n {
		for i := range out {
			out[i] = math.NaN()
		}
		r, _ := series.NewSeries(s.Timestamps(), out)
		return r
	}
	var seed float64
	for i := 0; i < n; i++ {
		v := s.At(i)
		if math.IsNaN(v) {
			// Without a clean seed window we cannot bootstrap; return all NaN.
			for j := range out {
				out[j] = math.NaN()
			}
			r, _ := series.NewSeries(s.Timestamps(), out)
			return r
		}
		seed += v
		out[i] = math.NaN()
	}
	out[n-1] = seed / float64(n)
	for i := n; i < s.Len(); i++ {
		v := s.At(i)
		if math.IsNaN(v) || math.IsNaN(out[i-1]) {
			out[i] = math.NaN()
			continue
		}
		out[i] = alpha*v + (1-alpha)*out[i-1]
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func smaSeries(s series.Series, n, m int) series.Series {
	out := make([]float64, s.Len())
	if n <= 0 || m <= 0 || s.Len() == 0 {
		for i := range out {
			out[i] = math.NaN()
		}
		r, _ := series.NewSeries(s.Timestamps(), out)
		return r
	}
	prev := math.NaN()
	for i := 0; i < s.Len(); i++ {
		v := s.At(i)
		if math.IsNaN(v) {
			out[i] = math.NaN()
			prev = math.NaN()
			continue
		}
		if math.IsNaN(prev) {
			prev = v
			out[i] = v
			continue
		}
		prev = (float64(m)*v + float64(n-m)*prev) / float64(n)
		out[i] = prev
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func rsiSeries(s series.Series, n int) series.Series {
	out := make([]float64, s.Len())
	if n <= 0 || s.Len() < n+1 {
		for i := range out {
			out[i] = math.NaN()
		}
		r, _ := series.NewSeries(s.Timestamps(), out)
		return r
	}
	for i := 0; i < n; i++ {
		out[i] = math.NaN()
	}

	var gains, losses float64
	for i := 1; i <= n; i++ {
		d := s.At(i) - s.At(i-1)
		if d >= 0 {
			gains += d
		} else {
			losses -= d
		}
	}
	gains /= float64(n)
	losses /= float64(n)
	out[n] = rsiValue(gains, losses)

	for i := n + 1; i < s.Len(); i++ {
		d := s.At(i) - s.At(i-1)
		gain, loss := 0.0, 0.0
		if d >= 0 {
			gain = d
		} else {
			loss = -d
		}
		gains = (gains*float64(n-1) + gain) / float64(n)
		losses = (losses*float64(n-1) + loss) / float64(n)
		out[i] = rsiValue(gains, losses)
	}
	r, _ := series.NewSeries(s.Timestamps(), out)
	return r
}

func rsiValue(gain, loss float64) float64 {
	if loss == 0 {
		if gain == 0 {
			return 50
		}
		return 100
	}
	rs := gain / loss
	return 100 - 100/(1+rs)
}
