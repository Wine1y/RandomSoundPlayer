package effects

import (
	"math/rand/v2"
	"time"

	"github.com/faiface/beep"
)

func GetRandomEffect() func(beep.Streamer, beep.Format) beep.Streamer {
	return Effects[rand.IntN(len(Effects))]
}

var Effects = []func(beep.Streamer, beep.Format) beep.Streamer{
	NewEchoEffect,
	NewBassEffect,
	NewSpeedupEffect,
	NewSlowdownEffect,
}

func NewEchoEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return NewEchoStreamer(streamer, format, time.Millisecond*time.Duration(rand.Int64N(250)+350), rand.IntN(4)+2)
}

func NewBassEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return NewBassStreamer(streamer, format, 0.025, rand.Float64()*40+10)
}

func NewSpeedupEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return beep.ResampleRatio(4, rand.Float64()+1, streamer)
}

func NewSlowdownEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return beep.ResampleRatio(4, rand.Float64()*0.5+0.5, streamer)
}
