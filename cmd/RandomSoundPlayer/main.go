package main

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Wine1y/RandomSoundPlayer/internal"
	"github.com/Wine1y/RandomSoundPlayer/internal/effects"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

const DEFAULT_SAMPLE_RATE = beep.SampleRate(44100)
const DEFAULT_SOUND_INTERVAL_SECONDS = 45.00
const DEFAULT_MULTICAST_DIVIDER = 5
const SOUND_PACKS_DIR = "sound_packs"

var soundCache = internal.NewSoundCache(25)

func loadSoundPacks() ([]internal.SoundPack, error) {
	packEntries, err := os.ReadDir(SOUND_PACKS_DIR)
	if err != nil {
		return make([]internal.SoundPack, 0), nil
	}

	var soundPacks []internal.SoundPack = make([]internal.SoundPack, 0, len(packEntries))
	for _, packEntry := range packEntries {
		if !packEntry.IsDir() {
			continue
		}
		packName := packEntry.Name()
		packPath := filepath.Join(SOUND_PACKS_DIR, packName)
		soundPack, err := internal.LoadSoundPackAt(packPath)
		if err != nil {
			slog.Warn("One of the sound packs wasn't loaded", "packPath", packPath, "err", err)
			continue
		}
		if len(soundPack.Sounds) > 0 {
			soundPacks = append(soundPacks, soundPack)
		}
	}
	return soundPacks, nil
}

func getSampleRate(soundPacks []internal.SoundPack) beep.SampleRate {
	sampleRates := make(map[beep.SampleRate]int, 8)
	for i := range soundPacks {
		for j := range soundPacks[i].Sounds {
			sampleRate := soundPacks[i].Sounds[j].Format.SampleRate

			_, exists := sampleRates[sampleRate]
			if exists {
				sampleRates[sampleRate]++
			} else {
				sampleRates[sampleRate] = 1
			}
		}
	}
	if len(sampleRates) == 0 {
		slog.Info("Using default sample rate", "sampleRate", DEFAULT_SAMPLE_RATE)
		return DEFAULT_SAMPLE_RATE
	}

	ratesList := make([]struct {
		sampleRate beep.SampleRate
		count      int
	}, 0, len(sampleRates))
	for sampleRate, count := range sampleRates {
		ratesList = append(ratesList, struct {
			sampleRate beep.SampleRate
			count      int
		}{sampleRate, count})
	}
	sort.Slice(ratesList, func(i, j int) bool {
		return ratesList[i].count > ratesList[j].count
	})

	var maxCountRates = 0
	for i := range ratesList {
		if ratesList[i].count == ratesList[0].count {
			maxCountRates++
		}
	}

	mostCommonRates := ratesList[:maxCountRates]
	sort.Slice(mostCommonRates, func(i, j int) bool {
		return mostCommonRates[i].sampleRate > mostCommonRates[j].sampleRate
	})

	sRate, sRateCount := mostCommonRates[len(mostCommonRates)/2].sampleRate, mostCommonRates[len(mostCommonRates)/2].count
	slog.Info("Using median most common sample rate", "sampleRate", sRate, "soundsInIt", sRateCount)
	return sRate
}

func getRandomSoundStreamer(
	soundPacks []internal.SoundPack,
	sampleRate beep.SampleRate,
) (*internal.SoundPack, *internal.Sound, beep.StreamSeeker) {
	pack := &soundPacks[rand.IntN(len(soundPacks))]
	sound := &pack.Sounds[rand.IntN(len(pack.Sounds))]

	buffer, cached := soundCache.Get(pack.Name, sound.Name)
	if !cached {
		buffer = beep.NewBuffer(beep.Format{
			SampleRate:  sampleRate,
			NumChannels: sound.Format.NumChannels,
			Precision:   sound.Format.Precision,
		})
		if sound.Format.SampleRate == sampleRate {
			buffer.Append(sound.Streamer)
		} else {
			buffer.Append(beep.Resample(4, sound.Format.SampleRate, sampleRate, sound.Streamer))
		}
		sound.Streamer.Seek(0)
		soundCache.Put(pack.Name, sound.Name, buffer)
	}
	return pack, sound, buffer.Streamer(0, buffer.Len())
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	soundPacks, err := loadSoundPacks()
	if err != nil {
		panic(err)
	}

	totalSounds, packStrings := 0, make([]string, 0, len(soundPacks))
	for i := range soundPacks {
		totalSounds += len(soundPacks[i].Sounds)
		packStrings = append(packStrings, fmt.Sprintf("%v (%v sounds)", soundPacks[i].Name, len(soundPacks[i].Sounds)))
	}
	slog.Info("Sound packs loaded",
		"total", fmt.Sprintf("%v packs, %v sounds", len(soundPacks), totalSounds),
		"packs", fmt.Sprintf("[%v]", strings.Join(packStrings, ", ")),
	)

	sampleRate := getSampleRate(soundPacks)

	mixer := beep.Mixer{}
	speaker.Init(sampleRate, int(time.Second/10*time.Duration(sampleRate)/time.Second))
	speaker.Play(&mixer)

	mean, deviation := DEFAULT_SOUND_INTERVAL_SECONDS, DEFAULT_SOUND_INTERVAL_SECONDS/2.5
	for {
		for i := 0.00; ; i++ {
			if rand.Float64() < 1.00/math.Pow(DEFAULT_MULTICAST_DIVIDER, i) {
				pack, sound, soundStreamer := getRandomSoundStreamer(soundPacks, sampleRate)
				slog.Debug("Playing sound", "packName", pack.Name, "soundName", sound.Name, "multicast", i)
				//TODO: Add chance to proc an effect instead of using it every time
				effect := effects.GetRandomEffect()
				mixer.Add(effect(soundStreamer, beep.Format{
					SampleRate:  sampleRate,
					NumChannels: sound.Format.NumChannels,
					Precision:   sound.Format.Precision,
				}))

				multicastPause := rand.Float64()*(float64(soundStreamer.Len())/float64(sampleRate)) + 0.02
				time.Sleep(time.Duration(float64(time.Second) * multicastPause))
			} else {
				break
			}
		}
		interval := time.Duration((math.Min(rand.NormFloat64(), 5)*deviation + mean) * float64(time.Second))
		slog.Debug("Next sound in", "interval", interval)
		time.Sleep(interval)
	}
}
