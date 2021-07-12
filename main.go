package main

import (
	"github.com/110y/run"

	"github.com/110y/bootes/internal/server"
)

func main() {
	run.Run(server.Run)
}
