package main

import (
	"github.com/rs/zerolog/log"
	"github.com/soldatov-s/go-garage-example/internal/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal().Err(err).Msg("run application")
	}
}
