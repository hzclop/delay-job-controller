package main

import (
	"fmt"
	app "k8s.io/delay-job-controller/pkg/server"
	"k8s.io/delay-job-controller/server"
	"os"
)

func main() {
	if err := app.Run(server.NewControllerServer()); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
