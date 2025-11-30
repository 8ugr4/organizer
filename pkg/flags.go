package pkg

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Flags struct {
	SrcPath  string
	DstPath  string
	RulePath string
	LogPath  string
	DryRun   bool
	Async    bool
	Verbose  bool
	Pattern  string
}

func GetFlags(args []string) Flags {
	srcPath := flag.String("src", "./testDir", "Source directory path")
	dstPath := flag.String("dst", "", "Destination directory path")
	rulePath := flag.String("rules", "./rules.yaml", "output category rules")
	log := flag.String("log", "", "Log path")
	dryRun := flag.Bool("dry-run", false, "Dry-run option")
	async := flag.Bool("async", false, "Faster async option, uses goroutines")
	verbose := flag.Bool("verbose", false, "Set to debug mode")
	pattern := flag.String("pattern", "", "image file pattern, e.g.: IMG_YEARMONTHDAY_HOURMINUTESECOND.ext, IMG_20220830_195427.jpg")
	// TODO: implement me: validate := flag.Bool("validate", false, "Enable SHA256 validation after copy operation")

	flag.Parse()

	if *srcPath == "" {
		fmt.Println("source path must be provided")
		flag.Usage()
		os.Exit(1)
	}

	if *dstPath == "" {
		*dstPath = strings.Join([]string{strings.TrimSuffix(*srcPath, "/"), "_cp"}, "")
		slog.Warn("destination path is not set by user", "auto-set destination path as", *dstPath)
	}

	if *verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if *rulePath == "" {
		slog.Warn("path for rules file is empty, going to use default settings from 'rules.yaml'")
		*rulePath = "./rules.yaml"
	}
	return Flags{
		SrcPath:  *srcPath,
		DstPath:  *dstPath,
		LogPath:  *log,
		DryRun:   *dryRun,
		Async:    *async,
		Verbose:  *verbose,
		RulePath: *rulePath,
		Pattern:  *pattern,
	}
	//TODO: separate img-sort and org-dir subcommand flag functions or structs. (find a better design method)
}

func GetSubCommand() string {
	if len(os.Args) < 2 {
		fmt.Println("expected 'org-dir' or 'sort-img' subcommand")
	}
	return os.Args[1]
}
