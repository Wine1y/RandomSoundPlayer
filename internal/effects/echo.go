package effects

import (
	"math"
	"time"

	"github.com/faiface/beep"
)

type EchoStreamer struct {
	streamer    beep.Streamer
	format      beep.Format
	waves       [][][2]float64
	samplesLeft int
}

func NewEchoStreamer(streamer beep.Streamer, format beep.Format, echoDelay time.Duration, wavesAmount int) *EchoStreamer {
	echo := EchoStreamer{
		streamer: streamer,
		format:   format,
		waves:    make([][][2]float64, 0, wavesAmount),
	}
	for i := range wavesAmount {
		echo.waves = append(echo.waves, make([][2]float64, format.SampleRate.N(echoDelay)*(i+1)))
	}
	echo.samplesLeft = len(echo.waves[len(echo.waves)-1])
	return &echo
}

func (echo *EchoStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if echo.samplesLeft <= 0 {
		return 0, false
	}

	n, ok = echo.streamer.Stream(samples)
	if !ok {
		echo.samplesLeft -= len(samples)
	}
	wavesOldest := make([][][2]float64, 0, len(echo.waves))
	for i := range echo.waves {
		waveOldest := make([][2]float64, len(samples))
		copy(waveOldest, echo.waves[i][:len(waveOldest)])
		copy(echo.waves[i], echo.waves[i][len(waveOldest):])
		copy(echo.waves[i][len(echo.waves[i])-len(waveOldest):], samples)
		wavesOldest = append(wavesOldest, waveOldest)
	}

	for i := range wavesOldest {
		for j := range samples {
			samples[j][0] += wavesOldest[i][j][0] * math.Pow(0.5, float64(i+1))
			samples[j][1] += wavesOldest[i][j][1] * math.Pow(0.5, float64(i+1))
		}
	}

	return len(samples), true
}

func (echo *EchoStreamer) Err() error {
	return echo.streamer.Err()
}
