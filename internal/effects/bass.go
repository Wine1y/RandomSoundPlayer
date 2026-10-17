package effects

import (
	"math"

	"github.com/faiface/beep"
)

type BassStreamer struct {
	streamer beep.Streamer
	alpha    float64
	boost    float64
}

func NewBassStreamer(streamer beep.Streamer, alpha float64, boost float64) BassStreamer {
	bass := BassStreamer{
		streamer: streamer,
		alpha:    alpha,
		boost:    boost,
	}
	return bass
}

func (bass BassStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	//It's not terribly designed if it's designed to be terrible
	// - Me, when i created this abomination
	n, ok = bass.streamer.Stream(samples)
	if ok {
		prev := [2]float64{0.00, 0.00}
		for i := range samples {
			prev[0] = prev[0] + bass.alpha*(samples[i][0]-prev[0])
			prev[1] = prev[1] + bass.alpha*(samples[i][1]-prev[1])

			samples[i][0] = math.Tanh(samples[i][0] + prev[0]*bass.boost)
			samples[i][1] = math.Tanh(samples[i][1] + prev[1]*bass.boost)
		}
	}
	return n, ok
}

func (bass BassStreamer) Err() error {
	return bass.streamer.Err()
}
