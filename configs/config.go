package configs

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type ConfigData struct {
	MasterVolume        float64    `toml:"master_volume"`
	SoundInterval       float64    `toml:"sound_interval"`
	MulticastBaseChance float64    `toml:"multicast_base_chance"`
	EffectChance        float64    `toml:"effect_chance"`
	SoundPacksDir       string     `toml:"sound_packs_dir"`
	SoundCacheSize      int        `toml:"sound_cache_size"`
	LoggingLevel        slog.Level `toml:"logging_level"`
	DisabledSoundPacks  []string   `toml:"disabled_sound_packs"`
	EffectWeights       map[string]struct {
		Weight int `toml:"weight"`
	} `toml:"effects"`
}

var Config ConfigData

const CONFIG_PATH = "config.toml"

const DEFAULT_MASTER_VOLUME = 1.00
const DEFAULT_SOUND_INTERVAL_SECONDS = 45.00
const DEFAULT_MULTICAST_BASE_CHANCE = 0.20
const DEFAULT_EFFECT_CHANCE = 0.15
const DEFAULT_SOUND_PACKS_DIR = "sound_packs"
const DEFAULT_SOUND_CACHE_SIZE = 25

func init() {
	Config = ConfigData{
		MasterVolume:        DEFAULT_MASTER_VOLUME,
		SoundInterval:       DEFAULT_SOUND_INTERVAL_SECONDS,
		MulticastBaseChance: DEFAULT_MULTICAST_BASE_CHANCE,
		EffectChance:        DEFAULT_EFFECT_CHANCE,
		SoundPacksDir:       DEFAULT_SOUND_PACKS_DIR,
		SoundCacheSize:      DEFAULT_SOUND_CACHE_SIZE,
		LoggingLevel:        slog.LevelInfo,
	}
	data, err := os.ReadFile(CONFIG_PATH)
	if !os.IsNotExist(err) {
		if err != nil {
			panic(err)
		}
		_, err = toml.Decode(string(data), &Config)
		if err != nil {
			panic(err)
		}
	}
}
