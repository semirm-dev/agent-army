package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the top-level configuration file for army.
type Config struct {
	Version int               `json:"version"`
	DirMap  map[string]string `json:"dir_map"`
}

// Path returns the fixed config path: ~/.army/config.json.
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".army", "config.json"), nil
}

// Load reads the config from ~/.army/config.json.
// Returns an empty config (version 1, empty map) if the file does not exist.
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	return LoadFrom(p)
}

// LoadFrom reads the config from the given path.
// Returns an empty config (version 1, empty map) if the file does not exist.
func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyConfig(), nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if cfg.DirMap == nil {
		cfg.DirMap = map[string]string{}
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}

	return &cfg, nil
}

// Save writes the config atomically to ~/.army/config.json (temp file + rename).
// Creates the parent directory if needed.
func Save(cfg *Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	return SaveTo(p, cfg)
}

// SaveTo writes the config atomically to the given path (temp file + rename).
// Creates the parent directory if needed.
func SaveTo(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "config-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}

	success = true
	return nil
}

// Register adds or updates a directory -> manifest mapping.
// Both dir and manifestPath are cleaned to absolute paths.
// If manifestPath equals defaultPath, the entry is removed instead.
func Register(cfg *Config, dir, manifestPath, defaultPath string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving directory: %w", err)
	}
	absDir = filepath.Clean(absDir)

	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("resolving manifest path: %w", err)
	}
	absManifest = filepath.Clean(absManifest)

	absDefault := filepath.Clean(defaultPath)
	if absManifest == absDefault {
		delete(cfg.DirMap, absDir)
		return nil
	}

	if cfg.DirMap == nil {
		cfg.DirMap = make(map[string]string)
	}
	cfg.DirMap[absDir] = absManifest
	return nil
}

// Resolve walks from the given directory upward to find a matching
// config entry. Returns the manifest path if found, or empty string if no match.
func Resolve(cfg *Config, dir string) string {
	dir = filepath.Clean(dir)
	for {
		if p, ok := cfg.DirMap[dir]; ok {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Remove deletes the mapping for the given directory.
func Remove(cfg *Config, dir string) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	delete(cfg.DirMap, filepath.Clean(absDir))
}

func emptyConfig() *Config {
	return &Config{
		Version: 1,
		DirMap:  map[string]string{},
	}
}
