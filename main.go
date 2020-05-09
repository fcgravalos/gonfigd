package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"

	"github.com/fcgravalos/gonfigd/gonfig"
	"github.com/rs/zerolog"
)

func main() {
	cfg := &gonfig.Config{}

	var fsWatchInterval time.Duration
	var enableDebugLog bool

	flag.StringVar(&cfg.GrpcAddr, "server-addr", ":8080", "gRPC server address.")
	flag.StringVar(&cfg.RootFolder, "root-folder", "./", "Root folder of the configuration tree")
	flag.DurationVar(&fsWatchInterval, "fswatch-interval", 5*time.Second, "How often the fswatcher will inspect the configuration tree to setup")
	flag.BoolVar(&enableDebugLog, "debug", false, "Enable debug logging")
	flag.Parse()

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
	cfg.KvKind = kv.INMEMORY
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
		case <-time.After(1 * time.Second):
			break
		}
	}
}
