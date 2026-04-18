package main

import (
	"backup_categorizer/pkg"
	"os"
	"time"
)

func main() {
	startTime := time.Now()

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

// TODO: check priv&public funcs
