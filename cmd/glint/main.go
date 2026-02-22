package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/nicholas/glint/pkg/config"
	"github.com/nicholas/glint/pkg/engine"
	"github.com/nicholas/glint/pkg/report"
	"github.com/nicholas/glint/pkg/rule"
	"github.com/spf13/cobra"

	// Register all rules via init()
	_ "github.com/nicholas/glint/pkg/rules/bugs"
	_ "github.com/nicholas/glint/pkg/rules/perf"
	_ "github.com/nicholas/glint/pkg/rules/security"
	_ "github.com/nicholas/glint/pkg/rules/style"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{
		Use:     "glint",
		Short:   "A super fast Go linter",
		Version: version,
	}

	root.AddCommand(runCmd())
	root.AddCommand(listRulesCmd())
	root.AddCommand(initConfigCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	var (
		configPath  string
		format      string
		enableAll   bool
		noCache     bool
		concurrency int
	)

	cmd := &cobra.Command{
		Use:   "run [packages...]",
		Short: "Run the linter on Go packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			var cfg *config.Config
			var err error
			if configPath != "" {
				cfg, err = config.LoadFile(configPath)
			} else {
				wd, _ := os.Getwd()
				cfg, err = config.Load(wd)
			}
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if format != "" {
				cfg.Output.Format = format
			}
			if enableAll {
				cfg.EnableAll = true
			}
			if noCache {
				cfg.Cache.Enabled = false
			}
			if concurrency > 0 {
				cfg.Concurrency = concurrency
			}

			eng, err := engine.New(cfg, rule.GlobalRegistry())
			if err != nil {
				return err
			}

			diags, err := eng.Run(context.Background(), args)
			if err != nil {
				return fmt.Errorf("analysis failed: %w", err)
			}

			reporter := report.New(cfg.Output.Format, cfg.Output.Color)
			if err := reporter.Report(os.Stdout, diags); err != nil {
				return fmt.Errorf("reporting: %w", err)
			}

			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "glint: analyzed %d package(s) with %d rule(s) in %s\n",
				len(args), len(eng.ActiveRules()), elapsed.Round(time.Millisecond))

			if len(diags) > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "path to config file")
	cmd.Flags().StringVarP(&format, "format", "f", "", "output format: text, json, sarif")
	cmd.Flags().BoolVar(&enableAll, "enable-all", false, "enable all rules regardless of config")
	cmd.Flags().BoolVar(&noCache, "no-cache", false, "disable result caching")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "j", 0, "number of concurrent workers (0 = NumCPU)")

	return cmd
}

func listRulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "List all available lint rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			rules := rule.GlobalRegistry().All()
			sort.Slice(rules, func(i, j int) bool {
				if rules[i].Category() != rules[j].Category() {
					return rules[i].Category() < rules[j].Category()
				}
				return rules[i].Name() < rules[j].Name()
			})

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "RULE\tCATEGORY\tSEVERITY\tTYPES\tDESCRIPTION\n")
			for _, r := range rules {
				needsTypes := "no"
				if r.NeedsTypeInfo() {
					needsTypes = "yes"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					r.Name(), r.Category(), r.Severity(), needsTypes, r.Description())
			}
			return tw.Flush()
		},
	}
}

func initConfigCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a default .glint.yml config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ".glint.yml"
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists; remove it first", path)
			}
			if err := config.WriteDefault(path); err != nil {
				return err
			}
			fmt.Printf("Created %s with default configuration.\n", path)
			return nil
		},
	}
}
