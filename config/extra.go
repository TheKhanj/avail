package config

import (
	"os"
	"path/filepath"

	"github.com/thekhanj/avail/common"
)

func (this *Config) GetPidFile() string {
	if this.PidFile != nil {
		return *this.PidFile
	}

	return filepath.Join(common.GetVarDir(), "main.pid")
}

func ReadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	err = c.UnmarshalJSON(b)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
