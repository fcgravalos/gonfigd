package fswatcher

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/fcgravalos/gonfigd/kv"
	"github.com/fcgravalos/gonfigd/pubsub"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	kv   kv.KV
	ps   pubsub.PubSub
	log  zerolog.Logger
	root string
}

var testCfg *testConfig

func TestUpsertFileOnDb(t *testing.T) {
	fullPath := fmt.Sprintf("%s/test.yaml", testCfg.root)
	err := ioutil.WriteFile(fullPath, []byte("foo: bar"), 0644)
	if err != nil {
		panic(err)
	}
	fsw := &fsWatcher{kv: testCfg.kv, ps: testCfg.ps}
	changed, e1 := fsw.upsertFileOnDb(fullPath)
	assert.Nil(t, e1)
	assert.True(t, changed)

	v1, e2 := testCfg.kv.Get(fmt.Sprintf("%s/test.yaml", testCfg.root))
	assert.Nil(t, e2)
	assert.Equal(t, "foo: bar", v1.Text())

	changed2, e3 := fsw.upsertFileOnDb(fullPath)
	assert.Nil(t, e3)
	assert.False(t, changed2)

	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString("\nbar: baz"); err != nil {
		panic(err)
	}
	changed3, e4 := fsw.upsertFileOnDb(fullPath)
	assert.Nil(t, e4)
	assert.True(t, changed3)

	v2, e5 := testCfg.kv.Get(fmt.Sprintf("%s/test.yaml", testCfg.root))
	assert.Nil(t, e5)
	assert.Equal(t, "foo: bar\nbar: baz", v2.Text())
}

func TestIsValidFileName(t *testing.T) {
	assert.True(t, isValidFileName("foo.yaml"))
	assert.False(t, isValidFileName("foo.swp"))
	assert.False(t, isValidFileName("foo.swx"))
	assert.False(t, isValidFileName("foo.~"))
	assert.False(t, isValidFileName("foo.tmp"))
}

func TestIsValidFile(t *testing.T) {
	assert.False(t, isValidFile("foo.yaml"))

	fp1 := fmt.Sprintf("%s/foo.yaml", testCfg.root)
	e1 := ioutil.WriteFile(fp1, []byte("foo: bar"), 0644)
	if e1 != nil {
		panic(e1)
	}
	assert.True(t, isValidFile(fmt.Sprintf("%s/foo.yaml", testCfg.root)))

	fp2 := fmt.Sprintf("%s/foo.swp", testCfg.root)
	e2 := ioutil.WriteFile(fp2, []byte("foo: bar"), 0644)
	if e2 != nil {
		panic(e2)
	}
	assert.False(t, isValidFile(fmt.Sprintf("%s/foo.swp", testCfg.root)))
}

func TestPublishEvent(t *testing.T) {
	fsw := &fsWatcher{ps: testCfg.ps}
	e1 := fsw.ps.CreateTopic("foo/config.yaml")
	assert.Nil(t, e1)

	sub, e2 := fsw.ps.Subscribe("foo/config.yaml")
	assert.Nil(t, e2)

	var ev *pubsub.Event

	done := make(chan struct{})

	go func(done chan struct{}) {
		scH := sub.Channel()
		ev = <-scH
		done <- struct{}{}
	}(done)

	e3 := fsw.publishEvent("foo/config.yaml", pubsub.ConfigCreated)
	assert.Nil(t, e3)
	<-done
	assert.Equal(t, "foo/config.yaml", ev.ConfigPath())
	assert.Equal(t, pubsub.ConfigCreated, ev.Kind())
}

func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fp := fmt.Sprintf("%s/test-start.yaml", testCfg.root)
	testCfg.ps.CreateTopic(fp)
	sub, _ := testCfg.ps.Subscribe(fp)
	sCh := sub.Channel()

	go Start(ctx, testCfg.root, 5*time.Second, testCfg.kv, testCfg.ps, testCfg.log)

	err := ioutil.WriteFile(fp, []byte("foo: bar"), 0644)
	if err != nil {
		panic(err)
	}

	ev1 := <-sCh
	assert.Equal(t, pubsub.ConfigCreated, ev1.Kind())

	v1, e1 := testCfg.kv.Get(fp)
	assert.Nil(t, e1)
	assert.Equal(t, "foo: bar", v1.Text())

	f, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err := f.WriteString("\nbar: baz"); err != nil {
		panic(err)
	}

	ev2 := <-sCh
	assert.Equal(t, pubsub.ConfigUpdated, ev2.Kind())

	v2, e2 := testCfg.kv.Get(fp)
	assert.Nil(t, e2)
	assert.Equal(t, "foo: bar\nbar: baz", v2.Text())

	os.Remove(fp)

	ev3 := <-sCh
	assert.Equal(t, pubsub.ConfigDeleted, ev3.Kind())

	v3, e3 := testCfg.kv.Get(fp)
	assert.Nil(t, v3)
	assert.True(t, kv.IsKeyNotFoundError(e3))
	assert.EqualError(t, e3, fmt.Sprintf("[%s] Key %s not found in KV", kv.KeyNotFound, fp))
}

func TestMain(m *testing.M) {
	dir, err := ioutil.TempDir("", fmt.Sprintf("fswatcher-tests-%s", time.Now()))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dir)

	kv, _ := kv.NewKV(kv.INMEMORY)
	ps, _ := pubsub.NewPubSub(pubsub.INMEMORY)

	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Caller().
		Logger()

	testCfg = &testConfig{
		kv:   kv,
		ps:   ps,
		log:  logger,
		root: dir,
	}
	os.Exit(m.Run())
}
