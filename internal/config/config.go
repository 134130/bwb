// Package config reads non-secret user configuration; it never writes credentials.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Profile struct {
	InsecureSkipReprompt bool   `toml:"insecure_skip_reprompt" json:"insecureSkipReprompt"`
	LockTimeout          string `toml:"lock_timeout" json:"lockTimeout"`
}

type File struct {
	Version        int                `toml:"version"`
	DefaultProfile string             `toml:"default_profile"`
	Profiles       map[string]Profile `toml:"profiles"`
}

var profileName = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

func Normalize(p Profile) (Profile, error) {
	if p.LockTimeout == "" {
		p.LockTimeout = "0s"
	}
	timeout, err := time.ParseDuration(p.LockTimeout)
	if err != nil || timeout < 0 {
		return Profile{}, errors.New("lock_timeout must be a nonnegative duration")
	}
	p.LockTimeout = timeout.String()
	return p, nil
}

// Path follows XDG on all supported platforms, including macOS.
func Path(flag string) (string, bool, error) {
	path := flag
	if path == "" {
		path = os.Getenv("BWB_CONFIG")
	}
	explicit := path != ""
	if !explicit {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return "", false, err
		}
	}
	absolute, err := filepath.Abs(path)
	return absolute, explicit, err
}

// DefaultPath ignores BWB_CONFIG and returns the optional default location.
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	if !filepath.IsAbs(base) {
		return "", errors.New("XDG_CONFIG_HOME must be absolute")
	}
	return filepath.Join(base, "bwb", "config.toml"), nil
}

func Load(path string, explicit bool) (fileConfig *File, loadErr error) {
	defer func() {
		if loadErr != nil {
			loadErr = fmt.Errorf("config %s: %w", path, loadErr)
		}
	}()
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) && !explicit {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open config: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil {
		return nil, errors.New("cannot read config")
	}
	if len(data) > 1<<20 {
		return nil, errors.New("config exceeds 1 MiB")
	}
	var config File
	if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&config); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) {
			for _, field := range strict.Errors {
				key := field.Key()
				if len(key) == 3 && key[0] == "profiles" && (key[2] == "server" || key[2] == "email") {
					return nil, fmt.Errorf("%s is no longer a config field; remove profiles.<name>.%s and enter it at login or use --%s", key[2], key[2], key[2])
				}
			}
			row, column := strict.Errors[0].Position()
			return nil, fmt.Errorf("unknown config key at line %d, column %d; see dev/config.example.toml for supported settings", row, column)
		}
		var decode *toml.DecodeError
		if errors.As(err, &decode) {
			row, column := decode.Position()
			return nil, fmt.Errorf("invalid TOML syntax or value type at line %d, column %d", row, column)
		}
		// Parser excerpts can contain accidentally supplied passwords or tokens.
		return nil, errors.New("invalid config: malformed TOML, unknown key, or invalid value type")
	}
	if config.Version == 0 {
		return nil, errors.New("config version is required; add version = 1 at the top of the file")
	}
	if config.Version != 1 {
		return nil, errors.New("unsupported config version; supported version is 1")
	}
	for name, p := range config.Profiles {
		if !profileName.MatchString(name) {
			return nil, errors.New("invalid config profile name")
		}
		normalized, err := Normalize(p)
		if err != nil {
			return nil, fmt.Errorf("profile %s: %w", name, err)
		}
		config.Profiles[name] = normalized
	}
	if config.DefaultProfile != "" {
		if _, ok := config.Profiles[config.DefaultProfile]; !ok {
			return nil, errors.New("default_profile does not exist")
		}
	}
	return &config, nil
}

func Select(flag string, changed bool, file *File) (string, error) {
	name := flag
	if !changed {
		name = os.Getenv("BWB_PROFILE")
		if name == "" && file != nil {
			name = file.DefaultProfile
		}
		if name == "" {
			name = "default"
		}
	}
	if file != nil {
		if _, ok := file.Profiles[name]; !ok {
			return "", errors.New("profile is not defined in config")
		}
	}
	return name, nil
}
