package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/GkadyrG/L2/L2.16/internal/mirror"
)

func main() {
	urlFlag := flag.String("url", "", "url to mirror")
	depth := flag.Int("depth", 1, "depth")
	out := flag.String("out", "./downloads", "output dir")
	workers := flag.Int("workers", 5, "workers")
	timeout := flag.Int("timeout", 30, "timeout sec")
	flag.Parse()

	if *urlFlag == "" {
		fmt.Println("missing -url")
		os.Exit(2)
	}

	opts := mirror.Options{
		RootURL:   *urlFlag,
		MaxDepth:  *depth,
		OutputDir: *out,
		Workers:   *workers,
		Timeout:   time.Duration(*timeout) * time.Second,
	}

	if err := mirror.Run(opts); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
