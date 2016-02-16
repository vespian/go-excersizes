package main

import (
	"bufio"
	"flag"
	"fmt"
	// "github.com/davecheney/profile"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const PACKET_LEN int = 5000
const BUF_PACKETS int = 10

var NUM_THREADS int

type FreqHash map[uint8]uint64

type SortedMap struct {
	Keys []uint8
	Map  FreqHash
}

func (a SortedMap) Len() int {
	return len(a.Map)
}

func (a SortedMap) Swap(i, j int) {
	a.Map[a.Keys[i]], a.Map[a.Keys[j]] = a.Map[a.Keys[j]], a.Map[a.Keys[i]]
}

func (a SortedMap) Less(i, j int) bool {
	return a.Map[a.Keys[i]] < a.Map[a.Keys[j]]
}

func breadcrumb2len(b string) uint8 {
	return uint8(len(strings.Split(b, "/")))
}

func fileParser(path string) (out chan []string, free_buffers chan []string) {
	// Take a file pointed by path argument, split it into PACKET_LEN lines
	// "packages" and send each one to "out" channel. In order to not to strain
	// GC to much, we reuse the buffers - only limited number of them is created
	// during start, later they are "returned" by goroutines processing the
	// them using "free_buffers" channel.

	//(sur) Is this model/approach optimal ?
	out = make(chan []string, BUF_PACKETS)
	free_buffers = make(chan []string, BUF_PACKETS)
	for i := 0; i < BUF_PACKETS-1; i++ {
		free_buffers <- make([]string, PACKET_LEN, PACKET_LEN)
	}
	var cur_buf []string = make([]string, PACKET_LEN, PACKET_LEN)

	go func() {
		var c int
		file, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if c < PACKET_LEN {
				cur_buf[c] = line
				c += 1
				continue
			} else {
				out <- cur_buf
				c = 0
				cur_buf = <-free_buffers
			}
		}

		// Send what is left in the buffer
		if c > 0 {
			for c < PACKET_LEN {
				cur_buf[c] = ""
				c += 1
			}
			out <- cur_buf
		}

		// Wait for processing subroutines to finish
		for i := 0; i < BUF_PACKETS; i++ {
			<-free_buffers
		}

		close(out)
		close(free_buffers)

		if err := scanner.Err(); err != nil {
			panic(fmt.Sprintf("%v", err))
		}
	}()

	return out, free_buffers
}

func parseBreadcrumbs(input_ch <-chan []string,
	free_buffers_ch chan<- []string) chan FreqHash {
	// Paring/regexpes is the most CPU-intensive task here, each go-routine
	// processes a buffer/package prepared by goroutines spawned by fileParser()
	// and then returns back the buffer.
	// The goroutine aggregates data about breadcrumbs/number of shashes locally
	// and when there is no more data, it sends it to output channel to the
	// goroutine that aggregates it into one big hash.
	var wg sync.WaitGroup
	out := make(chan FreqHash, BUF_PACKETS)

	body := func(gorun_number int) {
		breadcrumb_re := regexp.MustCompile(`^\s*<Topic\s+r:id=\"([^"]*)\">\s*$`)
		res := make(FreqHash)
		var t_processed int
		var t_counted int
		var num_bredc uint8

		for buf := range input_ch {
			for _, s := range buf {
				if s == "" {
					continue
				}
				t_processed += 1
				match := breadcrumb_re.FindStringSubmatch(s)
				if match == nil || len(match) < 2 {
					continue
				}
				num_bredc = breadcrumb2len(match[1])
				t_counted += 1
				res[num_bredc] += 1
			}
			free_buffers_ch <- buf
		}
		out <- res
		fmt.Printf("goroutine: %d - processed lines %d\n", gorun_number,
			t_processed)
		fmt.Printf("goroutine: %d - breadcrumbs found %d\n", gorun_number,
			t_counted)
		wg.Done()
	}

	for i := 0; i < NUM_THREADS; i++ {
		go body(i)
		wg.Add(1)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func aggregateBreadcrumbs(in <-chan FreqHash) FreqHash {
	// This functin aggregates the data produced by parseBreadcrumbs() goroutines
	// and outputs an aggregated map with all lengths found in the input file.
	res := make(FreqHash)
	var t uint64

	for c := range in {
		for key, val := range c {
			res[key] += val
			t += val
		}
	}
	fmt.Println("counted total: ", t)
	return res
}

func print_histogram(h FreqHash) {
	var tmp SortedMap

	tmp.Map = h
	tmp.Keys = make([]uint8, 0, len(h))

	for key, _ := range h {
		tmp.Keys = append(tmp.Keys, key)
	}

	sort.Sort(tmp)

	for _, d := range tmp.Keys {
		fmt.Printf("Length %2d: %8d\n", d, tmp.Map[d])
	}
}

func getAverageBreadcrumbLen(h FreqHash) float64 {
	var sum_total float64
	var sum_weighted float64

	for key, val := range h {
		sum_total += float64(val)
		sum_weighted += float64(key) * float64(val)
	}

	return (sum_weighted / sum_total)

}

func getCmdline() (path string) {
	flag.StringVar(&path, "path", "./structure.rdf.u8",
		"the path to file to process")
	flag.Parse()

	return
}

func main() {
	// cfg := profile.Config{
	// 	//  MemProfile:     true,
	// 	//  CPUProfile:     true,
	// 	BlockProfile:   true,
	// 	NoShutdownHook: true, // do not hook SIGINT
	// }
	// // p.Stop() must be called before the program exits to
	// // ensure profiling information is written to disk.
	// p := profile.Start(&cfg)
	// defer p.Stop()
	NUM_THREADS = runtime.NumCPU()

	input_file_path := getCmdline()

	runtime.GOMAXPROCS(NUM_THREADS)

	// Let's create a processing pipeline!
	// Please see functions' comments for detailed info.
	file_ch, free_buffers_ch := fileParser(input_file_path)
	bread_crumbs_ch := parseBreadcrumbs(file_ch, free_buffers_ch)
	breakcrumb_freq_hash := aggregateBreadcrumbs(bread_crumbs_ch)
	average_breadcrumb_len := getAverageBreadcrumbLen(breakcrumb_freq_hash)
	fmt.Printf("Average len: %.2f\n", average_breadcrumb_len)
	print_histogram(breakcrumb_freq_hash)
}
