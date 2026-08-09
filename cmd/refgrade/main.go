package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gohaisee/refgrade/internal/check"
	"github.com/gohaisee/refgrade/internal/detect"
	"github.com/gohaisee/refgrade/internal/engine"
	"github.com/gohaisee/refgrade/internal/i18n"
	"github.com/gohaisee/refgrade/internal/project"
	"github.com/gohaisee/refgrade/internal/refgradeconfig"
	"github.com/gohaisee/refgrade/internal/report"
)

const (
	exitOK    = 0
	exitFail  = 1
	exitError = 2
)

var lastExitCode = exitOK

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	lastExitCode = exitOK
	if len(args) == 0 {
		printUsage(os.Stderr)
		return exitError
	}
	switch args[0] {
	case "scan":
		return runScanCLI(args[1:])
	case "detect":
		return runDetectCLI(args[1:])
	case "explain":
		return runExplainCLI(args[1:])
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", args[0])
		printUsage(os.Stderr)
		return exitError
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: refgrade <command> [flags] [path]")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  scan     scan a go module for refactor readiness")
	fmt.Fprintln(w, "  detect   detect stack libraries in a go module")
	fmt.Fprintln(w, "  explain  print when/why/fix for a check id")
}

func runScanCLI(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	lang := fs.String("lang", "", "report language (en, ru)")
	format := fs.String("format", "text", "output format (text, markdown, json)")
	var output string
	fs.StringVar(&output, "output", "", "write report to file")
	fs.StringVar(&output, "o", "", "write report to file")
	rest, err := parseFlags(fs, args)
	if err != nil {
		return exitError
	}
	path := "."
	if len(rest) > 0 {
		path = rest[0]
	}
	code, err := runScan(context.Background(), path, *lang, *format, output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return code
}

func runDetectCLI(args []string) int {
	fs := flag.NewFlagSet("detect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	lang := fs.String("lang", "", "report language (en, ru)")
	rest, err := parseFlags(fs, args)
	if err != nil {
		return exitError
	}
	path := "."
	if len(rest) > 0 {
		path = rest[0]
	}
	if err := runDetect(context.Background(), path, *lang); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return exitOK
}

func runExplainCLI(args []string) int {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	lang := fs.String("lang", "", "report language (en, ru)")
	rest, err := parseFlags(fs, args)
	if err != nil {
		return exitError
	}
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "explain requires a check id")
		return exitError
	}
	path := "."
	if len(rest) > 1 {
		path = rest[1]
	}
	if err := runExplain(context.Background(), rest[0], path, *lang); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return exitOK
}

// parseFlags collects flags anywhere in args, then returns positional tail
func parseFlags(fs *flag.FlagSet, args []string) ([]string, error) {
	var flagArgs []string
	var positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "-" {
			positional = append(positional, a)
			continue
		}
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		if strings.Contains(a, "=") {
			flagArgs = append(flagArgs, a)
			continue
		}
		flagArgs = append(flagArgs, a)
		if needsValue(fs, a) && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			flagArgs = append(flagArgs, args[i+1])
			i++
		}
	}
	if err := fs.Parse(flagArgs); err != nil {
		return nil, err
	}
	return positional, nil
}

func needsValue(fs *flag.FlagSet, arg string) bool {
	name := strings.TrimLeft(arg, "-")
	if i := strings.Index(name, "="); i >= 0 {
		return false
	}
	f := fs.Lookup(name)
	if f == nil {
		return false
	}
	if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
		return false
	}
	return true
}

func runScan(ctx context.Context, path, flagLang, formatName, output string) (int, error) {
	modRoot, err := project.ModuleRoot(path)
	if err != nil {
		return exitError, err
	}
	lang, err := refgradeconfig.ResolveLang(flagLang, modRoot)
	if err != nil {
		return exitError, err
	}
	b, err := i18n.Load(lang)
	if err != nil {
		return exitError, err
	}
	format, err := report.ParseFormat(formatName)
	if err != nil {
		return exitError, err
	}

	mod, err := project.Load(ctx, path)
	if err != nil {
		return exitError, err
	}

	res, err := engine.Scan(ctx, mod, check.Catalog(), b.T)
	if err != nil {
		return exitError, err
	}

	out, err := report.Render(res, format, b)
	if err != nil {
		return exitError, err
	}
	if err := writeOutput(output, out); err != nil {
		return exitError, err
	}

	if engine.HasFail(res) {
		return exitFail, nil
	}
	return exitOK, nil
}

func runDetect(ctx context.Context, path, flagLang string) error {
	modRoot, err := project.ModuleRoot(path)
	if err != nil {
		return err
	}
	lang, err := refgradeconfig.ResolveLang(flagLang, modRoot)
	if err != nil {
		return err
	}
	b, err := i18n.Load(lang)
	if err != nil {
		return err
	}
	mod, err := project.Load(ctx, path)
	if err != nil {
		return err
	}
	stacks, err := detect.Detect(ctx, mod)
	if err != nil {
		return err
	}
	text := detect.FormatText(stacks, b.T("detect.title"), b.T("detect.none"))
	_, err = os.Stdout.WriteString(text)
	return err
}

func runExplain(ctx context.Context, checkID, path, flagLang string) error {
	_ = ctx
	modRoot, err := project.ModuleRoot(path)
	if err != nil {
		return err
	}
	lang, err := refgradeconfig.ResolveLang(flagLang, modRoot)
	if err != nil {
		return err
	}
	b, err := i18n.Load(lang)
	if err != nil {
		return err
	}
	when, why, fix, ok := check.Explain(checkID, b.T)
	if !ok {
		return fmt.Errorf("%s: %s", b.T("explain.unknown"), checkID)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "%s: %s\n", b.T("explain.title"), checkID)
	fmt.Fprintf(&out, "%s: %s\n", b.T("explain.when"), when)
	fmt.Fprintf(&out, "%s: %s\n", b.T("explain.why"), why)
	fmt.Fprintf(&out, "%s: %s\n", b.T("explain.fix"), fix)
	_, err = os.Stdout.WriteString(out.String())
	return err
}

func writeOutput(path string, data []byte) error {
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
