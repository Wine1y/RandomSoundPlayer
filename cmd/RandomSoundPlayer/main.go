package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
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

const SAMPLE_RATE = beep.SampleRate(44100)

var soundCache internal.SoundCache

func IsStdoutAvailable() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode()&(os.ModeCharDevice|os.ModeNamedPipe|os.ModeIrregular)) != 0 || stat.Size() >= 0
}

func init() {
	logToFile := flag.Bool("log", false, "Enable to save logs into the file")
	flag.Parse()

	var fileWriter io.Writer = io.Discard
	if *logToFile {
		logFile, err := os.OpenFile("logs.log", os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		fileWriter = logFile
	}

	var stdoutWriter io.Writer = io.Discard
	if IsStdoutAvailable() {
		stdoutWriter = os.Stdout
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(
		io.MultiWriter(fileWriter, stdoutWriter),
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
		if len(configs.Config.EnabledSoundPacks) == 0 {
			if slices.Contains(configs.Config.DisabledSoundPacks, packName) {
				continue
			}
		} else {
			if !slices.Contains(configs.Config.EnabledSoundPacks, packName) {
				continue
			}
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

func getRandomSoundStreamer(
	soundPacks []internal.SoundPack,
	totalSounds int,
) (*internal.SoundPack, *internal.Sound, beep.StreamSeeker, beep.Format) {
	var (
		pack  *internal.SoundPack
		sound *internal.Sound
	)
	soundI := rand.IntN(totalSounds)
	minSoundI := 0

	for i := range soundPacks {
		if soundI < minSoundI+len(soundPacks[i].Sounds) {
			pack = &soundPacks[i]
			sound = &soundPacks[i].Sounds[soundI-minSoundI]
			break
		}
		minSoundI += len(soundPacks[i].Sounds)
	}

	buffer, format, cached := soundCache.Get(string(*sound))
	if !cached {
		streamer, soundFormat, err := sound.Decode()
		if err != nil {
			slog.Warn("Error while decoding sound", "soundPath", string(*sound))
			return getRandomSoundStreamer(soundPacks, totalSounds)
		}
		format = &beep.Format{
			SampleRate:  SAMPLE_RATE,
			NumChannels: soundFormat.NumChannels,
			Precision:   soundFormat.Precision,
		}

		buffer = beep.NewBuffer(*format)
		if soundFormat.SampleRate == SAMPLE_RATE {
			buffer.Append(streamer)
		} else {
			buffer.Append(beep.Resample(4, soundFormat.SampleRate, SAMPLE_RATE, streamer))
		}
		streamer.Close()
		soundCache.Put(string(*sound), buffer, format)
	}
	return pack, sound, buffer.Streamer(0, buffer.Len()), *format
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

	mixer := beep.Mixer{}
	speaker.Init(SAMPLE_RATE, int(time.Second/10*time.Duration(SAMPLE_RATE)/time.Second))
	speaker.Play(&mixer)

	paused := false
	togglePause := make(chan bool)
	runSystray(togglePause, len(soundPacks), totalSounds)

	mean, deviation := configs.Config.SoundInterval, configs.Config.SoundInterval/2.5
	interval := time.Duration((math.Min(rand.NormFloat64(), 5)*deviation + mean) * float64(time.Second))
	var timer *time.Timer = time.NewTimer(interval)
	slog.Debug("First sound in", "interval", interval)
	for {
		select {
		case <-timer.C:
			if !paused {
				for i := 0.00; ; i++ {
					if rand.Float64() > math.Pow(configs.Config.MulticastBaseChance, i) {
						break
					}
					pack, sound, soundStreamer, format := getRandomSoundStreamer(soundPacks, totalSounds)
					var effectName string = "none"
					var streamer beep.Streamer = soundStreamer
					if rand.Float64() < configs.Config.EffectChance {
						name, effect := effects.GetRandomEffect()
						effectName = name
						streamer = effect(soundStreamer, format)
					}
					mixer.Add(internal.NewMasterVolumeStreamer(streamer, configs.Config.MasterVolume))
					slog.Debug("Playing sound", "packName", pack.Name, "soundName", sound.Name(), "multicast", i, "effect", effectName)

					multicastPause := rand.Float64()*(float64(soundStreamer.Len())/float64(SAMPLE_RATE)) + 0.02
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

func runSystray(togglePause chan bool, totalPacks int, totalSounds int) {
	onSystrayRun := func() {
		systray.SetTitle("RSP")
		systray.SetTooltip(fmt.Sprintf("RandomSoundPlayer [%v packs, %v sounds]", totalPacks, totalSounds))
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
