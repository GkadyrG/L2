package main

import (
	"fmt"
	"os"

	"github.com/beevik/ntp"
)

func main() {
	const host = "pool.ntp.org"

	currentTime, err := ntp.Time(host)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ntp error:", err)
		os.Exit(1)
	}

	fmt.Println(currentTime)
}
