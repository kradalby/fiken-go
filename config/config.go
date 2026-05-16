// Package config loads the fiken-go configuration via koanf, merging
// (in load order): file (default ~/.config/fiken/config.toml) →
// FIKEN_* env vars → flag overrides.
package config

import (
	"errors"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/toml/v2"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// Profile is the effective fiken-go configuration for one named
// account / company.
type Profile struct {
	Token   string `koanf:"token"`
	Company string `koanf:"company"`
	Lang    string `koanf:"lang"`
}

// Config is the parsed configuration file plus override layer.
type Config struct {
	DefaultProfile string             `koanf:"default_profile"`
	Profiles       map[string]Profile `koanf:"profiles"`
	envOverride    Profile
}

// Load reads filePath (if it exists) and FIKEN_* env vars and
// returns a Config. flagOverrides may be nil.
func Load(filePath string, flagOverrides map[string]string) (*Config, error) {
	k := koanf.New(".")
	if filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			if err := k.Load(file.Provider(filePath), toml.Parser()); err != nil {
				return nil, err
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, err
	}

	envK := koanf.New(".")
	_ = envK.Load(env.Provider(".", env.Opt{
		Prefix: "FIKEN_",
		TransformFunc: func(k, v string) (string, any) {
			return strings.ToLower(strings.TrimPrefix(k, "FIKEN_")), v
		},
	}), nil)
	cfg.envOverride = Profile{
		Token:   envK.String("token"),
		Company: envK.String("company"),
		Lang:    envK.String("lang"),
	}

	if envP := envK.String("profile"); envP != "" {
		cfg.DefaultProfile = envP
	}

	if flagOverrides != nil {
		if v := flagOverrides["profile"]; v != "" {
			cfg.DefaultProfile = v
		}
		if v := flagOverrides["token"]; v != "" {
			cfg.envOverride.Token = v
		}
		if v := flagOverrides["company"]; v != "" {
			cfg.envOverride.Company = v
		}
		if v := flagOverrides["lang"]; v != "" {
			cfg.envOverride.Lang = v
		}
	}

	if cfg.DefaultProfile == "" {
		cfg.DefaultProfile = "default"
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return cfg, nil
}

// Resolve returns the effective Profile for `name`. If name is empty
// the DefaultProfile is used. Env / flag overrides apply on top.
func (c *Config) Resolve(name string) (Profile, bool) {
	if name == "" {
		name = c.DefaultProfile
	}
	base, ok := c.Profiles[name]
	if !ok && name == "default" {
		base = Profile{}
		ok = true
	}
	if !ok {
		return Profile{}, false
	}
	merged := merge(base, c.envOverride)
	return merged, true
}

func merge(base, over Profile) Profile {
	if over.Token != "" {
		base.Token = over.Token
	}
	if over.Company != "" {
		base.Company = over.Company
	}
	if over.Lang != "" {
		base.Lang = over.Lang
	}
	return base
}
