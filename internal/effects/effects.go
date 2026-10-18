package effects

import (
	"math/rand/v2"
	"time"

	"github.com/Wine1y/RandomSoundPlayer/configs"
	"github.com/faiface/beep"
)

func GetRandomEffect() (string, func(beep.Streamer, beep.Format) beep.Streamer) {
	weightsSum := 0
	for i := range Effects {
		weightsSum += Effects[i].Weight
	}

	r := rand.IntN(weightsSum)
	for _, e := range Effects {
		if r < e.Weight {
			return e.Name, e.Constructor
		}
		r -= e.Weight
	}
	panic("Unreachable")
}

var Effects = []struct {
	Name        string
	Weight      int
	Constructor func(beep.Streamer, beep.Format) beep.Streamer
}{
	{Weight: 25, Name: "Speedup", Constructor: NewSpeedupEffect},
	{Weight: 25, Name: "Slowdown", Constructor: NewSlowdownEffect},
	{Weight: 20, Name: "AutoPanning", Constructor: NewAutoPanEffect},
	{Weight: 20, Name: "Echo", Constructor: NewEchoEffect},
	{Weight: 10, Name: "Reverb", Constructor: NewReverbEffect},
	{Weight: 1, Name: "Bassboost", Constructor: NewBassEffect},
}

func init() {
	for i := range Effects {
		configEffectEntry, exists := configs.Config.EffectWeights[Effects[i].Name]
		if exists {
			Effects[i].Weight = configEffectEntry.Weight
		}
	}
}

func NewEchoEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return NewEchoStreamer(streamer, format, time.Millisecond*time.Duration(rand.Int64N(250)+350), rand.IntN(4)+2)
}

func NewReverbEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return NewEchoStreamer(streamer, format, time.Millisecond*25, 5)
}

func NewBassEffect(streamer beep.Streamer, _ beep.Format) beep.Streamer {
	return NewBassStreamer(streamer, 0.025, rand.Float64()*40+10)
}

func NewSpeedupEffect(streamer beep.Streamer, _ beep.Format) beep.Streamer {
	return beep.ResampleRatio(4, rand.Float64()+1, streamer)
}

func NewSlowdownEffect(streamer beep.Streamer, _ beep.Format) beep.Streamer {
	return beep.ResampleRatio(4, rand.Float64()*0.5+0.5, streamer)
}

func NewAutoPanEffect(streamer beep.Streamer, format beep.Format) beep.Streamer {
	return NewAutoPanStreamer(streamer, format, rand.Float64()*2-1, rand.Float64() < 0.5, rand.Float64()*3.5+0.5)
}
