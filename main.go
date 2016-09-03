package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/honeycombio/libhoney-go"
	flag "github.com/jessevdk/go-flags"
)

var BuildID string

type Options struct {
	APIHost string `hidden:"true" long:"api_host" description:"APIHost for the Honeycomb API" default:"https://api.honeycomb.io/"`

	WriteKey string   `short:"k" long:"writekey" description:"Team write key" required:"true"`
	Dataset  string   `short:"d" long:"dataset" description:"Name of the dataset" required:"true"`
	Name     []string `short:"n" long:"name" description:"Metric name"`
	Val      []string `short:"v" long:"value" description:"Metric value"`
	Verbose  bool     `short:"V" long:"verbose" description:"Show output"`
}

func main() {
	var opts Options
	flagParser := flag.NewParser(&opts, flag.Default)
	if extraArgs, err := flagParser.Parse(); err != nil || len(extraArgs) != 0 {
		errAndExit("command line parsing error - call with --help for usage")
	}

	if len(opts.Name) != len(opts.Val) {
		errAndExit("Must have a value for each metric name - call with --help for usage")
	}

	c := libhoney.Config{
		WriteKey: opts.WriteKey,
		Dataset:  opts.Dataset,
		APIHost:  opts.APIHost,
	}
	libhoney.Init(c)
	defer libhoney.Close()

	ev := libhoney.NewEvent()
	for i, name := range opts.Name {
		if val, err := strconv.Atoi(opts.Val[i]); err == nil {
			ev.AddField(name, val)
		} else if val, err := strconv.ParseFloat(opts.Val[i], 64); err == nil {
			ev.AddField(name, val)
		} else {
			// add it as a string
			ev.AddField(name, opts.Val[i])
		}
	}
	if opts.Verbose {
		fmt.Println("sending event", ev)
	}
	ev.Send()
	rs := libhoney.Responses()
	rsp := <-rs
	if opts.Verbose {
		fmt.Printf("sent event %+v\n", map[string]interface{}{
			"status_code": rsp.StatusCode,
			"body":        strings.TrimSpace(string(rsp.Body)),
			"duration":    rsp.Duration,
			"error":       rsp.Err,
		})
	}

}

func errAndExit(reason string) {
	fmt.Printf("Error: %s\n", reason)
	os.Exit(1)
}
