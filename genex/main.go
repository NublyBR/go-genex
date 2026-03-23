package main

import (
	"bufio"
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

	argNum    int
	argSecure bool
)

func init() {
	cmd.Flags().IntVarP(&argNum, "num", "n", 8, "number of samples to generate; set to 0 to generate all possibilities in order")
	cmd.Flags().BoolVarP(&argSecure, "secure", "s", false, "use cryptographically secure PRNG")
}

func run(_ *cobra.Command, args []string) {
	opts := make([]genex.Option, 0, 1)

	if argSecure && argNum > 0 {
		opts = append(opts, genex.OptionRNG(genex.SecureRand))
	}

	start := time.Now()
	gen, err := genex.Compile(args[0], opts...)
	if err != nil {
		panic(err)
	}

	min, max := gen.Bounds()
	var bounds string
	if min == max {
		bounds = fmt.Sprint(min)
	} else {
		bounds = fmt.Sprintf("%d-%d", min, max)
	}
	fmt.Fprintf(os.Stderr, "] Compiled: %s\n", gen)
	fmt.Fprintf(os.Stderr, "] Count: \033[32m%s\033[0m | Bounds: \033[32m%s\033[0m | Complexity: \033[32m%d\033[0m\n",
		genex.Readable(gen.Count()),
		bounds,
		gen.Complexity(),
	)
	fmt.Fprintf(os.Stderr, "] Time: \033[32m%s\033[0m\n", time.Since(start))

	buf := bytes.NewBuffer(make([]byte, 0, max))
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	if argNum <= 0 {
		iter := gen.Iterate()
		for iter.Next() {
			iter.Get(buf)
			out.Write(buf.Bytes())
			out.WriteByte('\n')
			buf.Reset()
		}
	} else {
		for i := 0; i < argNum; i++ {
			gen.Sample(buf)
			out.Write(buf.Bytes())
			out.WriteByte('\n')
			buf.Reset()
		}
	}
}

func main() {
	cmd.Execute()
}
