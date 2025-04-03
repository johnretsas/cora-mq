package main

import (
	"benchmarking/dequeue"
	"benchmarking/enqueue"
	"flag"
	"fmt"
	"os"
)

func main() {
	enqueueFlag := flag.Bool("enqueue", false, "Run the enqueue test")
	dequeueFlag := flag.Bool("dequeue", false, "Run the dequeue test")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *helpFlag {
		fmt.Println("Usage of the program:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *enqueueFlag {
		enqueue.TestEnqueue()
		return
	}
	if *dequeueFlag {
		dequeue.TestDequeue()
		return
	} else {
		fmt.Println("No valid flag provided. Use -help to see available options.")
	}
}
