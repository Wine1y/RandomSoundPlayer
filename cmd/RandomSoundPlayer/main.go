package main

import (
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Wine1y/RandomSoundPlayer/assets"
	"github.com/Wine1y/RandomSoundPlayer/configs"
	"github.com/Wine1y/RandomSoundPlayer/internal"
	"github.com/Wine1y/RandomSoundPlayer/internal/effects"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/getlantern/systray"
)

var soundCache internal.SoundCache

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: configs.Config.LoggingLevel},
	)))
	slog.Debug("Config initialized", "config", configs.Config)

	soundCache = *internal.NewSoundCache(configs.Config.SoundCacheSize)
}

func loadSoundPacks() ([]internal.SoundPack, error) {
	packEntries, err := os.ReadDir(configs.Config.SoundPacksDir)
	if err != nil {
		return make([]internal.SoundPack, 0), nil
	}

	var soundPacks []internal.SoundPack = make([]internal.SoundPack, 0, len(packEntries))
	for _, packEntry := range packEntries {
		if !packEntry.IsDir() {
			continue
		}
		packName := packEntry.Name()
		if slices.Contains(configs.Config.DisabledSoundPacks, packName) {
			continue
		}
		packPath := filepath.Join(configs.Config.SoundPacksDir, packName)
		soundPack, err := internal.LoadSoundPackAt(packPath)
		if err != nil {
			slog.Warn("One of the sound packs wasn't loaded", "packPath", packPath, "err", err)
			continue
		}
		if len(soundPack.Sounds) > 0 {
			soundPacks = append(soundPacks, soundPack)
		} else {
			slog.Warn("Empty sound pack skipped", "packPath", packPath)
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
		panic("Unreachable, at least one sound should be loaded by now")
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
	soundPacks, err := loadSoundPacks()
	if err != nil {
		panic(err)
	}

	totalSounds, packStrings := 0, make([]string, 0, len(soundPacks))
	for i := range soundPacks {
		totalSounds += len(soundPacks[i].Sounds)
		packStrings = append(packStrings, fmt.Sprintf("%v (%v sounds)", soundPacks[i].Name, len(soundPacks[i].Sounds)))
	}
	if totalSounds == 0 {
		slog.Error(fmt.Sprintf(
			"At least one sound have to be loaded. Use %v directory to create sound packs and load sounds",
			configs.Config.SoundPacksDir,
		))
		os.Exit(1)
	}
	slog.Info("Sound packs loaded",
		"total", fmt.Sprintf("%v packs, %v sounds", len(soundPacks), totalSounds),
		"packs", fmt.Sprintf("[%v]", strings.Join(packStrings, ", ")),
	)

	sampleRate := getSampleRate(soundPacks)

	mixer := beep.Mixer{}
	speaker.Init(sampleRate, int(time.Second/10*time.Duration(sampleRate)/time.Second))
	speaker.Play(&mixer)

	paused := false
	togglePause := make(chan bool)
	runSystray(togglePause)

	mean, deviation := configs.Config.SoundInterval, configs.Config.SoundInterval/2.5
	var timer *time.Timer = time.NewTimer(time.Duration((math.Min(rand.NormFloat64(), 5)*deviation + mean) * float64(time.Second)))
	for {
		select {
		case <-timer.C:
			if !paused {
				for i := 0.00; ; i++ {
					if rand.Float64() > math.Pow(configs.Config.MulticastBaseChance, i) {
						break
					}
					pack, sound, soundStreamer := getRandomSoundStreamer(soundPacks, sampleRate)
					var effectName string = "none"
					var streamer beep.Streamer = soundStreamer
					if rand.Float64() < configs.Config.EffectChance {
						name, effect := effects.GetRandomEffect()
						effectName = name
						streamer = effect(soundStreamer, beep.Format{
							SampleRate:  sampleRate,
							NumChannels: sound.Format.NumChannels,
							Precision:   sound.Format.Precision,
						})
					}
					mixer.Add(internal.NewMasterVolumeStreamer(streamer, configs.Config.MasterVolume))
					slog.Debug("Playing sound", "packName", pack.Name, "soundName", sound.Name, "multicast", i, "effect", effectName)

					multicastPause := rand.Float64()*(float64(soundStreamer.Len())/float64(sampleRate)) + 0.02
					time.Sleep(time.Duration(float64(time.Second) * multicastPause))
				}
			}
			interval := time.Duration((math.Min(rand.NormFloat64(), 5)*deviation + mean) * float64(time.Second))
			timer = time.NewTimer(interval)
			slog.Debug("Next sound in", "interval", interval, "wasPaused", paused)
		case <-togglePause:
			paused = !paused
			togglePause <- paused
		}
	}
}

func runSystray(togglePause chan bool) {
	onSystrayRun := func() {
		systray.SetTitle("RSP")
		systray.SetTooltip("RandomSoundPlayer")
		systray.SetIcon(assets.TrayIconData)
		pauseButton := systray.AddMenuItem("Pause", "")
		go func() {
			for {
				<-pauseButton.ClickedCh
				togglePause <- true
				if <-togglePause {
					slog.Debug("Pausing playback")
					pauseButton.SetTitle("Resume")
				} else {
					slog.Debug("Resuming playback")
					pauseButton.SetTitle("Pause")
				}
			}
		}()

		quitButton := systray.AddMenuItem("Quit", "")
		go func() {
			<-quitButton.ClickedCh
			os.Exit(0)
		}()
	}

	go systray.Run(onSystrayRun, func() {})
}
