package gonfig

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fcgravalos/gonfigd/api"
	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var cfg *Config

// PickRandomTCPPort picks free TCP Port from localhost
func pickRandomTCPPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Fatalf("Could not resolve address: %v", err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Could not setup port %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestGetConfig(t *testing.T) {
	var conn *grpc.ClientConn
	conn, e1 := grpc.Dial(cfg.GrpcAddr, grpc.WithInsecure())
	assert.Nil(t, e1)
	assert.NotNil(t, conn)
	defer conn.Close()

	c := api.NewGonfigClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fp := fmt.Sprintf("%s/test.yaml", cfg.RootFolder)

	client, e2 := c.WatchConfig(ctx, &api.WatchConfigRequest{ConfigPath: fp})
	assert.Nil(t, e2)
	assert.NotNil(t, client)
	outCh := make(chan string)
	defer close(outCh)
	go func(outCh chan string) {
		for {
			select {
			case <-time.After(1 * time.Second):
				resp, _ := client.Recv()
				if resp != nil {
					outCh <- resp.GetEvent()
				}
			}
		}
	}(outCh)

	err := ioutil.WriteFile(fp, []byte("foo: bar"), 0644)
	if err != nil {
		panic(err)
	}
	ev := <-outCh
	assert.True(t, strings.Contains(ev, pubsub.ConfigCreated.String()))
	response, err := c.GetConfig(ctx, &api.GetConfigRequest{ConfigPath: fp})
	assert.Nil(t, err)
	assert.Equal(t, response.GetConfig(), "foo: bar")
}

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	waitChan := make(chan struct{})

	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Caller().
		Logger()

	dir, err := ioutil.TempDir("", fmt.Sprintf("gonfig-tests-%s", time.Now()))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dir)
	cfg = &Config{
		GrpcAddr:       fmt.Sprintf(":%d", pickRandomTCPPort()),
		KvKind:         kv.INMEMORY,
		PsKind:         pubsub.INMEMORY,
		RootFolder:     dir,
		FsWalkInterval: 5 * time.Second,
		Logger:         logger,
	}

	go Start(ctx, waitChan, *cfg)

	res := m.Run()
	cancel()
	<-waitChan
	os.Exit(res)
}
