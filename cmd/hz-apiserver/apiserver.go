package main

import (
	"os"

	"github.com/sunweiwe/horizon/cmd/hz-apiserver/app"
	"k8s.io/component-base/cli"
)

func main() {
	cmd := app.NewAPIServerCommand()
	code := cli.Run(cmd)
	os.Exit(code)
}
