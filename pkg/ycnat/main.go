package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	vpc "github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const usage = `
ycnat

Make a network interface to be a default route for a Yandex VPC route table.

Arguments:
  -help print this message and exist
`

const retry = 60

type args struct {
	rt *string
	ip *string
}

var argv = args{
	rt: flag.String("rt", "", "required. path to yandex vpc route table id"),
	ip: flag.String("ip", "", "required. path to local ipv4 address"),
}

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
	}

	log.SetOutput(os.Stdout)
	log.SetPrefix("ycnat [INF] ")
	log.SetFlags(0)
}

func run() error {
	if *argv.rt == "" {
		return errors.New("missing required flag -rt")
	}
	if *argv.ip == "" {
		return errors.New("missing required flag -ip")
	}

	rt, err := ioutil.ReadFile(*argv.rt)
	if err != nil {
		return fmt.Errorf("unable to read route table id: %w", err)
	}

	ip, err := ioutil.ReadFile(*argv.ip)
	if err != nil {
		return fmt.Errorf("unable to read local ipv4: %w", err)
	}

	log.Printf("route table id: %s", string(rt))
	log.Printf("local ipv4: %s", string(ip))

	ctx := context.Background()

	log.Println("initializing yandex cloud sdk")
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return err
	}

	log.Println("updating route table")
	op, err := sdk.VPC().RouteTable().Update(ctx, &vpc.UpdateRouteTableRequest{
		RouteTableId: string(rt),
		StaticRoutes: []*vpc.StaticRoute{
			{
				Destination: &vpc.StaticRoute_DestinationPrefix{
					DestinationPrefix: "0.0.0.0/0",
				},
				NextHop: &vpc.StaticRoute_NextHopAddress{
					NextHopAddress: string(ip),
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("unable to update route table id: %w", err)
	}

	for i := 0; i < retry && !op.Done; i++ {
		log.Println("waiting for update operation to be completed")
		time.Sleep(1 * time.Second)
	}

	log.Println("done")
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stdout, "ycnat [ERR]: %v\n", err)
		os.Exit(1)
	}
}
