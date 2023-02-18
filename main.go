package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/DavidGamba/go-getoptions"
	"github.com/prometheus-community/pro-bing"
)

var (
	Count      int
	HostFmt    string
	ICMP       bool
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

	// Give it an extra second.
	Timeout = time.Duration(Count+1) * time.Second

	err := outputHeader()
	if err != nil {
		die("%v", err)
	}

	var wg sync.WaitGroup
	for i, host := range hosts {
		wg.Add(1)
		go pingHost(host, uint64(i), &wg)
	}

	wg.Wait()
}

func parseArgs() []string {
	opt := getoptions.New()
	opt.Bool("help", false, opt.Alias("h", "?"))

	opt.BoolVar(&ICMP, "icmp", false, opt.Alias("i"),
		opt.Description("requires privileges"))
	opt.IntVar(&Count, "count", 30, opt.Alias("c"))
	opt.BoolVar(&OutputCSV, "output-csv", false, opt.Alias("csv"))

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

func pingHost(host string, tracker uint64, wg *sync.WaitGroup) {
	defer wg.Done()

	pinger, err := probing.NewPinger(host)
	if err != nil {
		warn("%s error: %v", host, err)
		return
	}

	// To use UDP ping (ICMP == false) on Linux:
	// # sysctl -w net.ipv4.ping_group_range="0 2147483647"

	// To use ICMP ping on Linux without root:
	// # setcap cap_net_raw=+ep ping-monitor

	pinger.SetPrivileged(ICMP) // Actually determines whether it uses ICMP or UDP
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

	HostFmt = fmt.Sprintf("%%-%ds", MaxHostLen)
	fmt.Printf(
		// host..  ###/###   ############
		HostFmt+"  Packets   Round trip times\n"+
			HostFmt+"  Received  %-12v %-12v %-12v %-12v\n",
		"", "Host", "Minimum", "Maximum", "Mean", "Std. Dev.")

	return nil
}

func outputStats(host string, stats *probing.Statistics) error {
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

	fmt.Printf(
		HostFmt+"  %3d/%-3d   %-12v %-12v %-12v %-12v\n",
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
