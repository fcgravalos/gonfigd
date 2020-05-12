package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"

	"github.com/fcgravalos/gonfigd/gonfig"
	"github.com/rs/zerolog"
)

// Will be initialized at build time
var version string

func main() {
	cfg := &gonfig.Config{}

	var enableDebugLog bool
	var kvImpl string
	var versionFlag bool

	flag.BoolVar(&versionFlag, "version", false, "Show gonfigd version")
	flag.StringVar(&cfg.GrpcAddr, "server-addr", ":8080", "gRPC server address.")
	flag.StringVar(&cfg.RootFolder, "root-folder", "./", "Root folder of the configuration tree")
	flag.StringVar(&kvImpl, "kv", "in-memory", "Key-Value implementation. Only 'in-memory' supported")
	flag.DurationVar(&cfg.FsWalkInterval, "fswalk-interval", 5*time.Second, "How often the fswatcher will inspect the configuration tree for new folders. Example: 10s")
	flag.BoolVar(&enableDebugLog, "debug", false, "Enable debug logging")
	flag.Parse()

	if versionFlag {
		fmt.Printf("gonfigd version %s\n", version)
		os.Exit(0)
	}

	// Setting logs
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Caller().
		Logger()

	if enableDebugLog {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	cfg.Logger = logger

	// Add a better way of selecting these, when we actually support more kvs and pubsubs.
	kvkind, err := kv.KVFromName(kvImpl)
	if err != nil {
		logger.Fatal().Msgf("%v", err)
	}
	cfg.KvKind = kvkind
	cfg.PsKind = pubsub.INMEMORY
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	waitChan := make(chan struct{}, 1)

	go func() {
		gonfig.Start(ctx, waitChan, *cfg)
	}()

	for {
		select {
		case s := <-sigChan:
			logger.Info().
				Msgf("received %s signal, gracefully stopping...", s.String())
			cancel()
			break
		case <-waitChan:
			logger.Info().
				Msg("gonfigd stopped, goodbye!")
			return
		}
	}
}
