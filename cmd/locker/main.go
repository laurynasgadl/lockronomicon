package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/laurynasgadl/lockronomicon/api"
	"github.com/laurynasgadl/lockronomicon/build"
	"github.com/laurynasgadl/lockronomicon/pkg/locker"
)

var (
	flagAddr string
	flagPath string
	flagVers bool
)

func init() {
	flag.StringVar(&flagAddr, "address", ":80", "Network address to listen on")
	flag.StringVar(&flagPath, "path", "/opt/locker", "FS locker workdir path")
	flag.BoolVar(&flagVers, "v", false, "Binary version")
	flag.Parse()
}

func main() {
	if flagVers {
		fmt.Printf("%s %s (%s %s)\n", build.Name, build.Version, build.Date, build.Revision)
		os.Exit(0)
	}

	locker, err := locker.NewFsLocker(flagPath)
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewServer(locker)

	log.Printf("Listening on %s\n", flagAddr)
	if err := server.ListenAndServe(flagAddr); err != nil {
		log.Fatal(err)
	}
}
