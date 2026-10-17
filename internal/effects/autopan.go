package effects

import (
	"math"
	"time"

	"github.com/faiface/beep"
)

type AutoPanStreamer struct {
	streamer              beep.Streamer
	factor                float64 // -1 - channel 1 only, 0 - both channels equally, 1 - channel 2 only
	factorIncreasing      bool    // true - going ltr, false - rtl
	factorChangePerSample float64 // how much factor changes per one sample
}

func NewAutoPanStreamer(streamer beep.Streamer, format beep.Format, factor float64, ltrDirection bool, factorChangePerSecond float64) *AutoPanStreamer {
	autopan := AutoPanStreamer{
		streamer:              streamer,
		factor:                factor,
		factorIncreasing:      ltrDirection,
		factorChangePerSample: factorChangePerSecond / float64(format.SampleRate.N(time.Second)),
	}
	return &autopan
}

func (autopan *AutoPanStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = autopan.streamer.Stream(samples)
	if ok {
		for i := range samples {
			dif := 1 - math.Abs(autopan.factor)
			if autopan.factor < 0 {
				samples[i][1] *= dif
			} else {
				samples[i][0] *= dif
			}

			if autopan.factorIncreasing {
				autopan.factor = math.Min(1, autopan.factor+autopan.factorChangePerSample)
				if autopan.factor == 1 {
					autopan.factorIncreasing = false
				}
			} else {
				autopan.factor = math.Max(-1, autopan.factor-autopan.factorChangePerSample)
				if autopan.factor == -1 {
					autopan.factorIncreasing = true
				}
			}
		}
	}
	return n, ok
}

func (autopan *AutoPanStreamer) Err() error {
	return autopan.streamer.Err()
}
