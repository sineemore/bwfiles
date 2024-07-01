package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type application struct {
	config        config
	rootDirectory string
	dryRun        bool
	allowDir      bool
	allowRemove   bool
}

func usage(err error) {
	code := 0
	out := os.Stdout

	if err != nil {
		code = 1
		out = os.Stderr
		fmt.Fprintf(out, "error: %v\n", err)
	}

	fmt.Fprintf(out, strings.TrimLeft(`
usage: %s [-dDRh] [-c config] [-s root] command

Params:
  -c config  Specifies a full path for the configuration file.
  -d         Dry run.
  -s root    Specifies a full path for the target root directory.
  -D         Allow creating directories.
             This option 
  -R         Allow removing items in bitwarden.
  -h         Show this help message.

Commands:
  sync    Sync local files to bitwarden items.
  unpack  Sync bitwarden items to local files.

Environment:
  BWFILES_CONFIG  Specifies a full path for the configuration file.
  BW_SESSION      Specifies a bitwarden session token.
                  Is not used directly by bwfiles, but may be used by the bitwarden CLI.

Configurations are loaded from the first found of the following locations:
 - The path specified by the -c option.
 - The path specified by the BWFILES_CONFIG environment variable.
 - $XDG_CONFIG_HOME/bwfiles/config.json
 - $HOME/.bwfilesrc
`, "\n"), os.Args[0])
	os.Exit(code)
}

func main() {
	var (
		configPath, rootDirectory     string
		dryRun, allowRemove, allowDir bool
	)

	args, err := parsearg(os.Args[1:], func(o rune, gets func() (string, error)) error {
		switch o {
		case 'c':
			v, err := gets()
			if err != nil {
				return err
			}
			configPath = v
		case 'd':
			dryRun = true
		case 's':
			v, err := gets()
			if err != nil {
				return err
			}
			rootDirectory = v
		case 'D':
			allowDir = true
		case 'R':
			allowRemove = true
		case 'h':
			usage(nil)
		default:
			return fmt.Errorf("unknown option: -%c", o)
		}
		return nil
	})
	if err != nil {
		usage(err)
	}

	if len(args) != 1 {
		usage(errors.New("specify a command"))
	}

	config, err := loadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	app := application{
		config:        config,
		rootDirectory: rootDirectory,
		dryRun:        dryRun,
		allowRemove:   allowRemove,
		allowDir:      allowDir,
	}

	switch args[0] {
	case "unpack":
		err = app.unpack()
	case "sync":
		err = app.sync()
	case "config":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		enc.SetEscapeHTML(false)
		err = enc.Encode(config)
	default:
		usage(errors.New("unknown command"))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
