package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thekhanj/avail/common"
)

func (this *Config) GetPidFile() string {
	if this.PidFile != nil {
		return *this.PidFile
	}

	return filepath.Join(common.GetBaseVarDir(), "main.pid")
}

func (this *Ping) GetCheck() (Check, error) {
	if this.Check == nil {
		return nil, nil
	}

	invalidErr := fmt.Errorf("Invalid check strategy: %v", this.Check)
	if m, ok := this.Check.(map[string]any); ok {
		b, err := json.Marshal(this.Check)
		if err != nil {
			return nil, err
		}
		switch m["type"] {
		case "shell":
			var c ShellCheck
			err = c.UnmarshalJSON(b)
			if err != nil {
				return nil, err
			}
			return &c, nil
		case "exec":
			var c ExecCheck
			err = c.UnmarshalJSON(b)
			if err != nil {
				return nil, err
			}
			return &c, nil
		}
	}

	return nil, invalidErr
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
