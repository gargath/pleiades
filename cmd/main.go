package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/gargath/pleiades/pkg/consumer"
)

func main() {
	log.Printf("Pleiades %s\n", version())

	viper.SetEnvPrefix("PLEIADES")
	viper.AutomaticEnv()

	//	flag.String("listenAddr", "0.0.0.0:8080", "address to listen on")
	flag.Bool("help", false, "print this help and exit")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	if viper.GetBool("help") {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, flag.CommandLine.FlagUsages())
		os.Exit(0)
	}

	c := &consumer.Consumer{}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		log.Println("Shutting down...")
		c.Stop()
	}()

	log.Printf("Starting to consume events\n")
	lastEventID, err := c.Start()
	if err != nil {
		log.Printf("Event consumer exited with error: %v", err)
	}
	log.Println("Shutdown complete")
	log.Printf("Last seen Event ID: %s", lastEventID)
}
