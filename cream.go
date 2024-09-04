package main

import (
	"gopkg.in/ini.v1"
)

type CreamConfig struct {
	Steam     CreamSteam        `ini:"steam"`
	SteamMisc CreamSteamMisc    `ini:"steam_misc"`
	Dlc       map[string]string `ini:"-"`
}

func NewCreamConfig() *CreamConfig {
	file, err := ini.Load("./log_build/linux/x64/cream_api.ini")
	if err != nil {
		panic(err)
	}
	var config CreamConfig
	err = file.MapTo(&config)
	if err != nil {
		panic(err)
	}
	sect, err := file.GetSection("dlc")
	if err != nil {
		panic(err)
	}
	config.Dlc = make(map[string]string)
	for _, key := range sect.Keys() {
		config.Dlc[key.Name()] = key.String()
	}
	return &config
}

func (c *CreamConfig) Save() {
	cfg := ini.Empty()
	err := cfg.ReflectFrom(&c)
	if err != nil {
		panic(err)
	}
	dlc, err := cfg.GetSection("dlc")
	if err != nil {
		panic(err)
	}
	for i, j := range c.Dlc {
		dlc.Key(i).SetValue(j)
	}
	err = cfg.SaveTo("./test.ini")
	if err != nil {
		panic(err)
	}
}

type CreamSteam struct {
	Appid        int    `ini:"appid"`
	UnlockAll    bool   `ini:"unlockall"`
	ForceOffline bool   `ini:"forceoffline"`
	OrgApi       string `ini:"orgapi"`
	OrgApi64     string `ini:"orgapi64"`
}

type CreamSteamMisc struct {
	DisableUserInterface bool `ini:"disableuserinterface"`
}
