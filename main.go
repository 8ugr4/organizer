package main

import (
	"backup_categorizer/pkg"
	"fmt"
	"os"
	"time"
)

func main() {
	startTime := time.Now()
	process(pkg.GetSubCommand(), startTime)
}

func process(subcommand string, startTime time.Time) {
	switch subcommand {
	case "org-dir":
		copyDirs(startTime)
	case "img-sort":
		// TODO implement me
	default:
		fmt.Println("invalid subcommand expected 'org-dir' or 'sort-img' subcommand")
		os.Exit(1)
	}
}

func copyDirs(startTime time.Time) {
	o, err := pkg.GetNewOperator()
	if err != nil {
		panic(err)
	}

	o.Flags = pkg.GetFlags(os.Args[3:])
	if err := pkg.ValidateDir(o.Flags.SrcPath); err != nil {
		panic(err)
	}

	rules, err := pkg.ReadCategories(o.Flags.RulePath)
	if err != nil {
		panic(err)
	}

	if err := o.CreateSubdirs(o.Flags.DstPath, rules.Rules); err != nil {
		panic(err)
	}
	o.BuildStorageMaps(rules)

	if o.Flags.LogPath != "" {
		o.CsvHandler, err = pkg.NewCSVLogger(o.Flags.LogPath)
		if err != nil {
			panic(err)
		}
	}

	extensions, err := o.Operate()
	if err != nil {
		panic(err)
	}

	pkg.ResultLog(extensions, o, startTime)
	if o.Storage.Exif != nil {
		defer func() {
			err := o.Storage.Exif.Close()
			if err != nil {
				panic(err)
			}
		}()
	}
}
