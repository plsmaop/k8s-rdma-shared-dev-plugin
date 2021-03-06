package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/Mellanox/k8s-rdma-shared-dev-plugin/pkg/resources"
)

var (
	version = "master@git"
	commit  = "unknown commit"
	date    = "unknown date"
)

func printVersionString() string {
	return fmt.Sprintf("k8s-rdma-shared-dev-plugin version:%s, commit:%s, date:%s", version, commit, date)
}

func main() {
	// Init command line flags to clear vendor packages' flags, especially in init()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// add version flag
	versionOpt := false
	flag.BoolVar(&versionOpt, "version", false, "Show application version")
	flag.BoolVar(&versionOpt, "v", false, "Show application version")
	flag.Parse()
	if versionOpt {
		fmt.Printf("%s\n", printVersionString())
		return
	}

	log.Println("Starting K8s RDMA Shared Device Plugin version=", version)

	rm := resources.NewResourceManager()

	log.Println("resource manager reading configs")
	if err := rm.ReadConfig(); err != nil {
		log.Fatalln(err.Error())
	}

	if err := rm.ValidateConfigs(); err != nil {
		log.Fatalf("Exiting.. one or more invalid configuration(s) given: %v", err)
	}

	log.Println("Initializing resource servers")
	if err := rm.InitServers(); err != nil {
		log.Fatalf("Error: initializing resource servers %v \n", err)
	}

	log.Println("Starting all servers...")
	if err := rm.StartAllServers(); err != nil {
		log.Fatalf("Error: starting resource servers %v\n", err.Error())
	}

	log.Println("All servers started.")

	log.Println("Listening for term signals")
	log.Println("Starting OS watcher.")
	signalsNotifier := resources.NewSignalNotifier(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigs := signalsNotifier.Notify()

	s := <-sigs
	switch s {
	case syscall.SIGHUP:
		log.Println("Received SIGHUP, restarting.")
		if err := rm.RestartAllServers(); err != nil {
			log.Fatalf("unable to restart server %v", err)
		}
	default:
		log.Printf("Received signal \"%v\", shutting down.", s)
		_ = rm.StopAllServers()
		return
	}
}
