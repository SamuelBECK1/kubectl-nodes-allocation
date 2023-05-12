package main
import (
     	"github.com/SamuelBECK1/kubectl-nodes-allocation/cmd"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"text/tabwriter"
)

var version = "0.1.0"

func main () {
	cmd.SetVersion(version)
        w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
        defer w.Flush()
	nodeAllocationCmd := cmd.NewNodeAllocationCommand(genericclioptions.IOStreams{In: os.Stdin, Out: w, ErrOut: w})
        if err := nodeAllocationCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
