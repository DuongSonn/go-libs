package _redis

import (
	"errors"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}
}

func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if c.Port == 0 {
		return errors.New("port is required")
	}
	if c.DB < 0 {
		return errors.New("db must be greater than or equal to 0")
	}
	return nil
}

func (c *Config) GetDNS() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}
