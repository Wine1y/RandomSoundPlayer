package internal

import "github.com/faiface/beep"

type MasterVolumeStreamer struct {
	streamer beep.Streamer
	volume   float64
}

func NewMasterVolumeStreamer(streamer beep.Streamer, volume float64) MasterVolumeStreamer {
	return MasterVolumeStreamer{
		streamer: streamer,
		volume:   volume,
	}
}

func (masterVolume MasterVolumeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = masterVolume.streamer.Stream(samples)
	if ok && masterVolume.volume != 1 {
		for i := range samples {
			samples[i][0] = samples[i][0] * masterVolume.volume
			samples[i][1] = samples[i][1] * masterVolume.volume
		}
	}
	return n, ok
}

func (masterVolume MasterVolumeStreamer) Err() error {
	return masterVolume.streamer.Err()
}
