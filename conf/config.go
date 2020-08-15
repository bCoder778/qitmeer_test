package conf

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bCoder778/log"
	"sync"
	"time"
)

const (
	configFile = "config.toml"
)

var Setting *Config
var once sync.Once

func init() {
	once.Do(func() {
		_, err := toml.DecodeFile(configFile, &Setting)
		if err != nil {
			fmt.Printf("decode %s failed!, err:%s\n", configFile, err.Error())
		}
		decodeStart()
	})
}

type Config struct {
	Email       `toml:"email"`
	Log         `toml:"log"`
	Check       `toml:"check"`
	Task        `toml:"task"`
	ReleaseNode Node `toml:"releasenode"`
	TestNode    Node `toml:"testnode"`
}

type Email struct {
	User string   `toml:"user"`
	Pass string   `toml:"pass"`
	Host string   `toml:"host"`
	Port string   `toml:"port"`
	To   []string `toml:"to"`
}

type Log struct {
	Mode  log.Mode  `toml:"mode"`
	Level log.Level `toml:"level"`
}

type Node struct {
	Host string `toml:"host"`
	User string `toml:"user"`
	Pass string `toml:"pass"`
}

type Check struct {
	Order uint64 `toml:"order"`
}

type Task struct {
	Start     string `toml:"start"`
	Interval  int64  `toml:"interval"`
	Timestamp int64
}

func decodeStart() {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", Setting.Start, time.Local)
	if err != nil {
		fmt.Printf("decode start %s failed!, err:%s\n", Setting.Start, err.Error())
	}
	Setting.Timestamp = t.Unix()
}
