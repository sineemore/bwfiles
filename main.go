package main

import (
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

	dropPrivileges bool
	uid, gid       uint32
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
usage: %s [-dDRh] [-c config] [-s path] [-x user] command

Params:
  -d         Dry run.
  -D         Allow creating directories.
  -R         Allow removing items in bitwarden.
  -h         Show this help message.
  -c config  Specifies a full path for the configuration file.
  -s path    Specifies prefix for all files.
  -x user    Specifies the identity to run as.
             To allow running as root, specify "root".

Commands:
  sync    Sync local files to bitwarden items.
  unpack  Sync bitwarden items to local files.

Environment:
  BWFILES_CONFIG  Specifies a full path for the configuration file.
  BW_SESSION      Specifies a bitwarden session token.
                  Is not used directly by bwfiles, but may be used by the bitwarden CLI.
                  If running with sudo, try adding -E flag to preserve environment variables.

Configurations are loaded from the first found of the following locations:
 - The path specified by the -c option.
 - The path specified by the BWFILES_CONFIG environment variable.
 - $XDG_CONFIG_HOME/bwfiles/config.json
 - $HOME/.bwfilesrc
`, "\n"), os.Args[0])
	os.Exit(code)
}

func main() {
	var err error

	var (
		configPath, rootDirectory     string
		dryRun, allowRemove, allowDir bool

		allowRunningAsRoot bool
		identity           string
	)

	args, err := parsearg(os.Args[1:], func(o rune, gets func() (string, error)) error {
		switch o {
		case 'c':
			configPath, err = gets()
			if err != nil {
				return err
			}
		case 'd':
			dryRun = true
		case 's':
			rootDirectory, err = gets()
			if err != nil {
				return err
			}
		case 'D':
			allowDir = true
		case 'R':
			allowRemove = true
		case 'h':
			usage(nil)
		case 'x':
			identity, err = gets()
			if err != nil {
				return err
			}

			allowRunningAsRoot = identity == "root"
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

	isRoot, dropPrivileges, uid, gid, err := checkRunningAsRoot(identity, allowRunningAsRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if os.Getenv("BW_SESSION") == "" {
		fmt.Fprintln(os.Stderr, "warning: BW_SESSION is not set, bw may not work properly")
		if isRoot && os.Getenv("SUDO_USER") != "" {
			fmt.Fprintln(os.Stderr, "warning: you are running as root with sudo, try adding -E flag to preserve environment variables")
		}
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

		dropPrivileges: dropPrivileges,
		uid:            uid,
		gid:            gid,
	}

	switch args[0] {
	case "unpack":
		err = app.unpack()
	case "sync":
		err = app.sync()
	default:
		usage(errors.New("unknown command"))
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
