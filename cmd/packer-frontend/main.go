package main

import (
	"github.com/EricHripko/pack.yaml/internal/app/packer-frontend/cmd"
	_ "github.com/EricHripko/pack.yaml/pkg/plugins/golang"
)

func main() {
	cmd.Main()
}
