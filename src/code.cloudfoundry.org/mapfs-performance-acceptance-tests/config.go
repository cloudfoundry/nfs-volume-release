package perf

import (
	"encoding/json"
	"os"
)

type Config struct {
	Driver    string `json:"driver"`
	NfsServer string `json:"nfs_server"`
}

func LoadConfig() (Config, error) {
	fileName, avail := os.LookupEnv("CONFIG")
	if !avail {
		panic("CONFIG not set")
	}

	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	c := Config{}
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
