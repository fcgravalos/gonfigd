package gonfig

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/fcgravalos/gonfigd/api"
	"github.com/fcgravalos/gonfigd/fswatcher"
	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Config struct {
	GrpcAddr       string
	KvKind         kv.Kind
	PsKind         pubsub.Kind
	RootFolder     string
	FsWalkInterval time.Duration
	Logger         zerolog.Logger
}

func Start(ctx context.Context, waitChan chan struct{}, cfg Config) error {
	// create a server instance
	kv, err := kv.NewKV(cfg.KvKind)
	if err != nil {
		log.Fatalf("failed to create new kv instance: %v", err)
		return err
	}

	ps, err := pubsub.NewPubSub(cfg.PsKind)
	if err != nil {
		log.Fatalf("failed to create new pubsub instance: %v", err)
		return err
	}

	var wg sync.WaitGroup

	// Start fsWatcher
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		cfg.Logger.Info().
			Msg("starting fswatcher")
		if err := fswatcher.Start(ctx, cfg.RootFolder, cfg.FsWalkInterval, kv, ps, cfg.Logger); err != nil {
			cfg.Logger.Fatal().Msgf("fswatcher returned with error: %v", err)
		}
	}(ctx)

	// Start GRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf("%s", cfg.GrpcAddr))
	if err != nil {
		cfg.Logger.Error().Msgf("failed to listen at addr %s: %v", cfg.GrpcAddr, err)
		return err
	}

	s := api.NewServer(kv, ps, cfg.Logger)
	grpcServer := grpc.NewServer()
	api.RegisterGonfigServer(grpcServer, s)

	wg.Add(1)
	go func() {
		defer wg.Done()
		cfg.Logger.Info().
			Msg("starting gonfigd gRPC server")
		if err = grpcServer.Serve(lis); err != nil {
			cfg.Logger.Fatal().Msgf("failed to serve: %v", err)
		}
	}()

	<-ctx.Done()
	grpcServer.Stop()

	wg.Wait()
	waitChan <- struct{}{}
	return nil
}
