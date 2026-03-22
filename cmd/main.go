package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	genex "github.com/NublyBR/go-genex"

	"github.com/spf13/cobra"
)

var (
	cmd = &cobra.Command{
		Use:  "genex [pattern]",
		Args: cobra.ExactArgs(1),
		Run:  run,
	}

	samples int
)

func init() {
	cmd.Flags().IntVarP(&samples, "num", "n", 8, "number of samples to generate")
}

func run(_ *cobra.Command, args []string) {
	start := time.Now()

	g, err := genex.Compile(args[0])
	if err != nil {
		panic(err)
	}

	min, max := g.Bounds()
	var bounds string
	if min == max {
		bounds = fmt.Sprint(min)
	} else {
		bounds = fmt.Sprintf("%d-%d", min, max)
	}
	fmt.Fprintf(os.Stderr, "] Compiled: %s\n", g)
	fmt.Fprintf(os.Stderr, "] Count: \033[32m%s\033[0m | Bounds: \033[32m%s\033[0m | Complexity: \033[32m%d\033[0m\n",
		genex.Readable(g.Count()),
		bounds,
		g.Complexity(),
	)
	fmt.Fprintf(os.Stderr, "] Time: \033[32m%s\033[0m\n", time.Since(start))

	buf := bytes.NewBuffer(make([]byte, 0, max))

	for i := 0; i < samples; i++ {
		g.Sample(buf)
		fmt.Println(buf.String())
		buf.Reset()
	}
}

func main() {
	cmd.Execute()
}
