package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/sparrc/go-ping"
)

var (
	OutputCSV = true
	Count int
	Timeout time.Duration
	Interval time.Duration
)

func main() {
	var err error

	Count = 30

	Timeout, err = time.ParseDuration("60s")
	if err != nil {
		die("%v", err)
	}

	Interval, err = time.ParseDuration("1s")
	if err != nil {
		die("%v", err)
	}

	for _, host := range os.Args[1:] {
		err = pingHost(host)
		if err != nil {
			warn("%s ping failed: %v", host, err)
		}
	}
}

func pingHost(host string) error {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return err
	}

	pinger.Count = 30
	pinger.Timeout = Timeout
	pinger.Interval = Interval
	pinger.Run()

	return outputStats(host, pinger.Statistics())
}

func outputStats(host string, stats *ping.Statistics) error {
	if OutputCSV {
		w := csv.NewWriter(os.Stdout)

		err := w.Write([]string{
			host,
			fmt.Sprint(stats.PacketsRecv),
			fmt.Sprint(stats.PacketsSent),
			fmt.Sprintf("%.6f", stats.MinRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.MaxRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.AvgRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.StdDevRtt.Seconds()*1000),
		})
		if err != nil {
			return err
		}

		w.Flush()

		err = w.Error()
		if err != nil {
			return err
		}

		return nil
	}

	fmt.Printf(
		"%s  %d/%d success  min=%v  max=%v  avg=%v  stddev=%v\n",
		host, stats.PacketsRecv, stats.PacketsSent, stats.MinRtt, stats.MaxRtt,
		stats.AvgRtt, stats.StdDevRtt)

	return nil
}

func warn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg + "\n", args...)
}

func die(msg string, args ...interface{}) {
	warn(msg, args...)
	os.Exit(1)
}
