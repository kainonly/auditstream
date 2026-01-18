package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/kainonly/auditstream/v3/app"
	"github.com/kainonly/auditstream/v3/bootstrap"
	"github.com/kainonly/auditstream/v3/common"
)

func main() {
	var err error
	if common.Log, err = bootstrap.SetZap(); err != nil {
		panic(err)
	}

	values, err := bootstrap.LoadStaticValues("./config/values.yml")
	if err != nil {
		panic(err)
	}
	nc, err := bootstrap.UseNats(values)
	if err != nil {
		panic(err)
	}
	js, err := bootstrap.UseJetStream(nc)
	if err != nil {
		panic(err)
	}
	kv, err := bootstrap.UseKeyValue(values, js)
	if err != nil {
		panic(err)
	}
	schedule, err := bootstrap.UseSchedule()
	if err != nil {
		panic(err)
	}

	x := app.New(values, nc, js, kv, schedule)

	if err = x.States(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	if err = x.Run(ctx); err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
