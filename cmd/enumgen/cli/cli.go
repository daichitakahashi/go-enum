package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/daichitakahashi/go-enum/cmd/enumgen/gen"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:  "generate enum for Go",
	RunE: run,
}

var (
	wd           string
	out          string
	visitors     []string
	accepts      []string
	visitorImpls []string
)

func init() {
	flags := rootCmd.Flags()
	flags.StringVar(&wd, "wd", ".", "working directory")
	flags.StringVar(&out, "out", "enum.gen.go", "output file name")
	flags.StringSliceVar(&visitors, "visitor", nil, "")
	flags.StringSliceVar(&accepts, "accept", nil, "")
	flags.StringSliceVar(&visitorImpls, "visitor-impl", nil, "")
}

func run(cmd *cobra.Command, args []string) error {
	namingVisitorParams := make([]gen.NamingVisitorParams, 0, len(visitors))
	for _, v := range visitors {
		params, err := parseNamingVisitorParams(v)
		if err != nil {
			log.Fatalf("visitor: %s", err)
		}
		namingVisitorParams = append(namingVisitorParams, *params)
	}
	namingAcceptParams := make([]gen.NamingAcceptParams, 0, len(accepts))
	for _, a := range accepts {
		params, err := parseNamingAcceptParams(a)
		if err != nil {
			log.Fatalf("accept: %s", err)
		}
		namingAcceptParams = append(namingAcceptParams, *params)
	}
	namingVisitorImplParams := make([]gen.NamingVisitorImplParams, 0, len(visitorImpls))
	for _, f := range visitorImpls {
		namingVisitorImplParams = append(namingVisitorImplParams, parseNamingVisitorFactoryParams(f))
	}
	gen.Run(wd, out, namingVisitorParams, namingAcceptParams, namingVisitorImplParams)
	return nil
}

// --visitor="*Event:*Handler:On*"
func parseNamingVisitorParams(s string) (*gen.NamingVisitorParams, error) {
	parts := strings.SplitN(s, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid format %q", s)
	}
	return &gen.NamingVisitorParams{
		Target:     parts[0],
		TypeName:   parts[1],
		MethodName: parts[2],
	}, nil
}

// --accept="*Event:Emit"
func parseNamingAcceptParams(s string) (*gen.NamingAcceptParams, error) {
	target, name, ok := strings.Cut(s, ":")
	if !ok {
		return nil, fmt.Errorf("invalid format %q", s)
	}
	return &gen.NamingAcceptParams{
		Target:     target,
		MethodName: name,
	}, nil
}

// --visitor-factory="*Event"
// --visitor-factory="*Event:New*"
func parseNamingVisitorFactoryParams(s string) gen.NamingVisitorImplParams {
	target, name, ok := strings.Cut(s, ":")
	if !ok {
		name = "New*"
	}
	return gen.NamingVisitorImplParams{
		Target:      target,
		FactoryName: name,
	}
}

func Run() {
	_ = rootCmd.Execute()
}
