package config

import (
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/global"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"unsafe"
)

type Configuration struct {
	Database struct {
		Postgres struct {
			Host     string `koanf:"host"`
			Port     int    `koanf:"port"`
			User     string `koanf:"user"`
			Password string `koanf:"password"`
			Database string `koanf:"database"`
		} `koanf:"postgres"`
	} `koanf:"database"`

	RabbitMQ struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		User     string `koanf:"user"`
		Password string `koanf:"password"`
		Exchange string `koanf:"exchange"`
	} `koanf:"rabbitmq"`

	Logger struct {
		Level         string `koanf:"level"`
		File          string `koanf:"file"`
		Format        string `koanf:"format"`
		Output        string `koanf:"output"`
		IsDevelopment bool   `koanf:"is_development"`
		Encoding      string `koanf:"encoding"`
	} `koanf:"logger"`

	Backoff struct {
		InitialInterval int     `koanf:"initial_interval"`
		MaxInterval     int     `koanf:"max_interval"`
		Mulltiplier     float64 `koanf:"multiplier"`
		MaxElapsedTime  int     `koanf:"max_elapsed_time"`
	} `koanf:"backoff"`
}

var (
	// Use an unsafe pointer to hold the configuration for atomic swaps
	configPtr unsafe.Pointer
	k         = koanf.New(".")
)

func buildConfigPath() (path string, err error) {
	if global.IsDevelopment {
		return "./internal/config/config.yaml", nil
	}
	ex, err := os.Executable()
	if err != nil {
		return
	}
	exPath := filepath.Dir(ex)
	path = filepath.Join(exPath, "..", "..", "internal", "config", "config.yaml")
	return
}

func LoadConfiguration() error {
	path, err := buildConfigPath()
	if err != nil {
		return err
	}

	y := file.Provider(path)
	if err := k.Load(y, yaml.Parser()); err != nil {
		return err
	}
	if err := unmarshalIntoStruct(); err != nil {
		return err
	}
	log.Println("Configuration loaded successfully")
	return nil
}

// FileWatcher watches the file and gets a callback on change. The callback can do whatever,
// like re-load the configuration.
// File provider always returns a nil `event`.
func FileWatcher() {
	path, err := buildConfigPath()
	if err != nil {
		return
	}

	y := file.Provider(path)
	if err := y.Watch(func(event interface{}, err error) {
		if err != nil {
			log.Printf("watch error: %v", err)
		}

		// Throw away the old config and load a fresh copy.
		log.Println("config changed. Reloading ...")
		k = koanf.New(".")
		if err := k.Load(y, yaml.Parser()); err != nil {
			log.Printf("error loading config: %v", err)
		}
		if err := unmarshalIntoStruct(); err != nil {
			log.Printf("error unmarshaling config: %v", err)
		}
		log.Println("config reloaded.")

	}); err != nil {
		log.Printf("error watching config: %v", err)
	}
	log.Println("Config reloader running. Change config/config.yaml to reload the config.")
	<-global.CancellationContext().Done()
	log.Println("Config reloader stopped.")
}

func unmarshalIntoStruct() error {
	var params Configuration
	if err := k.UnmarshalWithConf("", &params, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
		return err
	}
	atomic.StorePointer(&configPtr, unsafe.Pointer(&params))
	return nil
}

func GetConfig() *Configuration {
	return (*Configuration)(atomic.LoadPointer(&configPtr))
}
