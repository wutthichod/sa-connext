package main

import (
	"log"

	"github.com/wutthichod/sa-connext/services/user-service/cmd"
)

func main() {

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
