package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

type args struct {
	rtid *string
}

var argv = args{
	rtid: flag.String("rtid", "", "required. yandex vpc route table id"),
}

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
	}
}

func run() error {
	if *argv.rtid == "" {
		return errors.New("missing required flag -rtid")
	}

	fmt.Println("getting local IP address from the instance metadata...")
	ipreq, err := http.Get("http://169.254.169.254/latest/meta-data/local-ipv4")
	if err != nil {
		return err
	}
	defer ipreq.Body.Close()

	ip, err := ioutil.ReadAll(ipreq.Body)
	if err != nil {
		return err
	}
	fmt.Printf("IP address is %s\n", string(ip))

	ctx := context.Background()

	fmt.Println("initializing yandex cloud sdk...")
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.InstanceServiceAccount(),
	})
	if err != nil {
		return err
	}

	fmt.Printf("updating route table %s\n", *argv.rtid)
	op, err := sdk.VPC().RouteTable().Update(ctx, &vpc.UpdateRouteTableRequest{
		RouteTableId: *argv.rtid,
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

	for !op.Done {
		fmt.Println("waiting for update operation to be completed...")
		time.Sleep(1 * time.Second)
	}

	fmt.Println("done")
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
