package pkg

import (
	"fmt"
	"path/filepath"

	"github.com/rumenvasilev/rvsecret/internal/config"
	"github.com/rumenvasilev/rvsecret/internal/log"
	"github.com/rumenvasilev/rvsecret/internal/pkg/api"
	"github.com/rumenvasilev/rvsecret/internal/util"
	"gopkg.in/yaml.v3"
)

func SaveConfig(log *log.Logger) error {
	log.Info("Creating configuration file")
	// implementation here
	// load config
	cfg, err := config.Load(api.Github)
	if err != nil {
		return err
	}

	if cfg.Global.Debug {
		log.Debug(config.PrintDebug("unknown"))
	}

	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(bytes))

	//homedir
	dir, err := util.MakeHomeDir(filepath.Dir(cfg.Global.ConfigFile), log)
	if err != nil {
		return err
	}

	path := dir + "/" + filepath.Base(cfg.Global.ConfigFile)
	//write
	log.Info("Writing to configuration file %s", path)
	return util.WriteToFile(path, bytes)
}
