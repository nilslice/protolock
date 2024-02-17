package protolock

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	LockDir   string
	ProtoRoot string
	Ignore    string
	UpToDate  bool
	Debug     bool
	Includes  []string
}

func NewConfig(
	lockDir, protoRoot, ignores string,
	upToDate, debug bool, includes string,
) (*Config, error) {
	l, err := filepath.Abs(lockDir)
	if err != nil {
		return nil, err
	}
	p, err := filepath.Abs(protoRoot)
	if err != nil {
		return nil, err
	}

	var includesAbs []string
	if len(includes) > 0 {
		includesOrig := strings.Split(includes, ",")
		for _, c := range includesOrig {
			i, err := filepath.Abs(c)
			if err != nil {
				return nil, err
			}
			includesAbs = append(includesAbs, i)
		}
	}

	return &Config{
		LockDir:   l,
		ProtoRoot: p,
		Ignore:    ignores,
		UpToDate:  upToDate,
		Includes:  includesAbs,
	}, nil
}

func (cfg *Config) LockFileExists() bool {
	_, err := os.Stat(cfg.LockFilePath())
	return err == nil && !os.IsNotExist(err)
}

func (cfg *Config) LockFilePath() string {
	return filepath.Join(cfg.LockDir, LockFileName)
}
