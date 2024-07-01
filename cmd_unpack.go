package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

func (app application) unpack() error {
	raw, err := app.getBitwardenItems()
	if err != nil {
		return err
	}

	var existing bwitems
	err = json.Unmarshal(raw, &existing)
	if err != nil {
		return err
	}

	items, err := app.bitwardenwToEntries(existing)
	if err != nil {
		return err
	}

	for _, item := range items {
		uid, gid, err := getIdentity(item.owner)
		if err != nil {
			return err
		}

		f, err := os.Open(item.path)
		if err == nil {
			defer f.Close()

			info, err := f.Stat()
			if err != nil {
				return err
			}

			content, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			f.Close()

			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				return fmt.Errorf("could not get owner of %s", item.path)
			}

			if bytes.Equal(content, item.content) &&
				uint32(info.Mode()) == item.mode &&
				stat.Uid == uid &&
				stat.Gid == gid {
				// Exactly the same file, skippping
				continue
			}
		}

		fmt.Fprintf(os.Stderr, "unpacking: %s\n", item.path)

		if app.dryRun {
			continue
		}

		if app.allowDir {
			err := os.MkdirAll(filepath.Dir(item.path), 0755)
			if err != nil {
				return err
			}
		}

		f, err = os.OpenFile(item.path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.FileMode(item.mode))
		if err != nil {
			var pe *os.PathError
			if errors.As(err, &pe) {
				fmt.Fprintf(os.Stderr, "TIP: to create directories, use -D option\n")
			}
			return err
		}
		defer f.Close()

		_, err = f.Write(item.content)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

		err = os.Chown(item.path, int(uid), int(gid))
		if err != nil {
			return err
		}
	}

	return nil
}
