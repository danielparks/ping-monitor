package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DavidGamba/go-getoptions"
	"github.com/sparrc/go-ping"
)

var (
	Count      int
	MaxHostLen = 1
	OutputCSV  bool
	Timeout    time.Duration
)

const Interval = 1 * time.Second

func main() {
	hosts := parseArgs()

	if !OutputCSV {
		for _, host := range hosts {
			if len(host) > MaxHostLen {
				MaxHostLen = len(host)
			}
		}
	}

	Timeout = time.Duration(2*Count) * time.Second

	err := outputHeader()
	if err != nil {
		die("%v", err)
	}

	// Pinger generates a tracking number for each instance, but if the instances
	// are created around the same time, they get the same number. This ensures
	// each pinger has a unique (but related) tracking number.
	trackerBase := time.Now().UnixNano()

	var wg sync.WaitGroup
	for i, host := range hosts {
		wg.Add(1)
		go pingHost(host, trackerBase+int64(i), &wg)
	}

	wg.Wait()
}

func parseArgs() []string {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("h", "?"))

	opt.BoolVar(&OutputCSV, "output-csv", false, opt.Alias("csv"))
	opt.IntVar(&Count, "count", 30, opt.Alias("c"))

	opt.SetMode(getoptions.Bundling) // -opt == -o -p -t
	opt.SetRequireOrder()            // stop processing after the first argument is found

	hosts, err := opt.Parse(os.Args[1:])
	if err != nil {
		warn("Error parsing arguments: %v", err)
		fmt.Fprintf(os.Stderr, opt.Help())
		os.Exit(1)
	}

	if opt.Called("help") {
		fmt.Print(opt.Help())
		os.Exit(0)
	}

	return hosts
}

func pingHost(host string, tracker int64, wg *sync.WaitGroup) {
	defer wg.Done()

	pinger, err := ping.NewPinger(host)
	if err != nil {
		warn("%s error: %v", host, err)
		return
	}

	pinger.Tracker = tracker
	pinger.Count = Count
	pinger.Timeout = Timeout
	pinger.Interval = Interval
	pinger.Run()

	err = outputStats(host, pinger.Statistics())
	if err != nil {
		warn("outputting stats: %v", err)
	}
}

func writeCSV(values ...string) error {
	w := csv.NewWriter(os.Stdout)

	err := w.Write(values)
	if err != nil {
		return err
	}

	w.Flush()

	return w.Error()
}

func outputHeader() error {
	if OutputCSV {
		return writeCSV(
			"host",
			"received",
			"sent",
			"min_ms",
			"max_ms",
			"mean_ms",
			"stddev_ms",
		)
	}

	return nil
}

func outputStats(host string, stats *ping.Statistics) error {
	if OutputCSV {
		return writeCSV(
			host,
			fmt.Sprint(stats.PacketsRecv),
			fmt.Sprint(stats.PacketsSent),
			fmt.Sprintf("%.6f", stats.MinRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.MaxRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.AvgRtt.Seconds()*1000),
			fmt.Sprintf("%.6f", stats.StdDevRtt.Seconds()*1000),
		)
	}

	hostFmt := fmt.Sprintf("%%-%ds", MaxHostLen)
	fmt.Printf(
		hostFmt+"  %3d/%-3d received  min=%-12v  max=%-12v  avg=%-12v  stddev=%-12v\n",
		host, stats.PacketsRecv, stats.PacketsSent, stats.MinRtt, stats.MaxRtt,
		stats.AvgRtt, stats.StdDevRtt)

	return nil
}

func warn(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
}

func die(msg string, args ...interface{}) {
	warn(msg, args...)
	os.Exit(1)
}
