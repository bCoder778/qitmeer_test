package main

import (
	"github.com/bCoder778/log"
	"github.com/bCoder778/qitmeer_test/conf"
	"github.com/bCoder778/qitmeer_test/test"
	"github.com/bCoder778/qitmeer_test/timer"
	"os"
	"os/signal"
	"sync"
)

func main() {
	log.SetOption(&log.Option{
		LogLevel: conf.Setting.Log.Level,
		Mode:     conf.Setting.Log.Mode,
		Email: &log.EMailOption{
			User:   conf.Setting.User,
			Pass:   conf.Setting.Pass,
			Host:   conf.Setting.Host,
			Port:   conf.Setting.Port,
			Target: conf.Setting.To,
		},
	})

	t := timer.New()
	t.Start(test.TestQitmeer, conf.Setting.Timestamp, conf.Setting.Interval)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_ = <-c
		t.Stop()
		wg.Done()
	}()
	wg.Wait()
}
