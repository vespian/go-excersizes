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

const packetLen int = 5000
const bufPackets int = 10

var numThreads int

type freqHash map[uint8]uint64

type sortedMap struct {
	Keys []uint8
	Map  freqHash
}

func (a sortedMap) Len() int {
	return len(a.Map)
}

func (a sortedMap) Swap(i, j int) {
	a.Map[a.Keys[i]], a.Map[a.Keys[j]] = a.Map[a.Keys[j]], a.Map[a.Keys[i]]
}

func (a sortedMap) Less(i, j int) bool {
	return a.Map[a.Keys[i]] < a.Map[a.Keys[j]]
}

func breadcrumbToLen(b string) uint8 {
	return uint8(len(strings.Split(b, "/")))
}

// Function fileParser takes a file pointed by path argument, splits it into
// packetLen lines "packages" and sends each one to "out" channel. In order
// to not to strain GC to much, we reuse the buffers - only limited number
// of them is created during start, later they are "returned" by goroutines
// processing the them using "freeBuffers" channel.
func fileParser(
	path string,
	errch chan<- error,
) (
	out chan []string,
	freeBuffers chan []string,
) {
	out = make(chan []string, bufPackets)
	freeBuffers = make(chan []string, bufPackets)

	for i := 0; i < bufPackets; i++ {
		freeBuffers <- make([]string, packetLen, packetLen)
	}

	go func() {
		var err error
		var file *os.File
		var c int
		curBuf := <-freeBuffers

		defer func() {
			if err != nil {
				errch <- err
			} else {
				if err = file.Close(); err != nil {
					errch <- err
				}
			}
			close(out)
		}()

		if file, err = os.Open(path); err != nil {
			err = fmt.Errorf("failed to open breadc file for "+
				"reading: %s", err)
			return
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if line := scanner.Text(); line != "" {
				curBuf[c] = line
				if c < packetLen-1 {
					c++
				} else {
					out <- curBuf
					curBuf = <-freeBuffers
					c = 0
				}
			}
		}

		// Send what is left in the buffer
		if c > 0 {
			curBuf[c] = ""
			out <- curBuf
		}

		if err = scanner.Err(); err != nil {
			return
		}
	}()

	return out, freeBuffers
}

// parseBreadcrumbs takes care of the most CPU-intensive task here: regexp
// matching. Each go-routine processes a buffer/package prepared by goroutines
// spawned by fileParser() and then returns back the buffer.
// The goroutine aggregates data about breadcrumbs/number of shashes locally
// and when there is no more data, it sends it to output channel to the
// goroutine that aggregates it into one big hash.
func parseBreadcrumbs(inputCh <-chan []string,
	freeBuffersCh chan<- []string) chan freqHash {

	var wg sync.WaitGroup
	out := make(chan freqHash, bufPackets)

	body := func(gorun_number int) {
		breadcrumbRe := regexp.MustCompile(`^\s*<Topic\s+r:id=\"([^"]*)\">\s*$`)
		res := make(freqHash)
		var tProcessed int
		var tCounted int
		var numBredc uint8

		for buf := range inputCh {
			for _, s := range buf {
				if s == "" {
					break
				}
				tProcessed++
				match := breadcrumbRe.FindStringSubmatch(s)
				if match == nil || len(match) < 2 {
					continue
				}
				numBredc = breadcrumbToLen(match[1])
				tCounted++
				res[numBredc]++
			}
			freeBuffersCh <- buf
		}
		out <- res
		fmt.Printf("goroutine: %d - processed lines %d\n", gorun_number,
			tProcessed)
		fmt.Printf("goroutine: %d - breadcrumbs found %d\n", gorun_number,
			tCounted)
		wg.Done()
	}

	for i := 0; i < numThreads; i++ {
		go body(i)
		wg.Add(1)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// Function aggregateBreadcrumbs aggregates the data produced by parseBreadcrumbs()
// goroutines and outputs an aggregated map with all lengths found in the input
// file.
func aggregateBreadcrumbs(in <-chan freqHash) chan freqHash {
	resCh := make(chan freqHash)

	go func() {
		res := make(freqHash)
		var t uint64

		for c := range in {
			for key, val := range c {
				res[key] += val
				t += val
			}
		}
		fmt.Println("counted total: ", t)
		resCh <- res
	}()

	return resCh
}

func printHistogram(h freqHash) {
	var tmp sortedMap

	tmp.Map = h
	tmp.Keys = make([]uint8, 0, len(h))

	for key := range h {
		tmp.Keys = append(tmp.Keys, key)
	}

	sort.Sort(tmp)

	for _, d := range tmp.Keys {
		fmt.Printf("Length %2d: %8d\n", d, tmp.Map[d])
	}
}

func getAverageBreadcrumbLen(h freqHash) float64 {
	var sumTotal float64
	var sumWeighted float64

	for key, val := range h {
		sumTotal += float64(val)
		sumWeighted += float64(key) * float64(val)
	}

	return (sumWeighted / sumTotal)

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
	numThreads = runtime.NumCPU()
	errch := make(chan error)
	var breakcrumbFreqHash freqHash

	inputFilePath := getCmdline()

	runtime.GOMAXPROCS(numThreads)

	// Let's create a processing pipeline!
	// Please see functions' comments for detailed info.
	fileCh, freeBuffersCh := fileParser(inputFilePath, errch)
	breadCrumbsCh := parseBreadcrumbs(fileCh, freeBuffersCh)
	breakcrumbFreqHashCh := aggregateBreadcrumbs(breadCrumbsCh)

	select {
	case breakcrumbFreqHash = <-breakcrumbFreqHashCh:
		break
	case err := <-errch:
		fmt.Fprintf(os.Stderr, "Error occured: %s", err)
		os.Exit(1)
	}

	averageBreadcrumbLen := getAverageBreadcrumbLen(breakcrumbFreqHash)
	fmt.Printf("Average len: %.2f\n", averageBreadcrumbLen)
	printHistogram(breakcrumbFreqHash)
}
