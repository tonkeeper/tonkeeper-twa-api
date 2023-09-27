package main

import (
	"log"
	"reflect"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	API struct {
		Port int `env:"PORT" envDefault:"7077"`
	}
	App struct {
		LogLevel    string `env:"LOG_LEVEL" envDefault:"ERROR"`
		PostgresURI string `env:"POSTGRES_URI,required"`
	}
	TonAPI struct {
		ApiKey string `env:"TONAPI_KEY,required"`
	}
	TonConnect struct {
		Secret string `env:"TON_CONNECT_SECRET,required"`
	}
	Telegram struct {
		BotSecretKey string `env:"TELEGRAM_BOT_SECRET_KEY,required"`
	}
}

func Load() Config {
	var c Config
	if err := env.ParseWithFuncs(&c, map[reflect.Type]env.ParserFunc{}); err != nil {
		log.Panicf("[‼️  Config parsing failed] %+v\n", err)
	}
	return c
}
