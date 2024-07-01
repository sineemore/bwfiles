package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func (app application) sync() error {
	files, err := app.selectFiles()
	if err != nil {
		return err
	}

	var entries []entry
	for _, file := range files {
		f, err := os.Open(file.path)
		if err != nil {
			return err
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		content, err := io.ReadAll(f)
		if err != nil {
			return err
		}

		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("could not get owner of %s", file.path)
		}

		u, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
		if err != nil {
			return err
		}

		g, err := user.LookupGroupId(strconv.Itoa(int(stat.Gid)))
		if err != nil {
			return err
		}

		entries = append(entries, entry{
			name:    file.name,
			path:    file.path,
			mode:    uint32(info.Mode()),
			owner:   u.Username + ":" + g.Name,
			content: content,
		})
	}
	raw, err := app.getBitwardenItems()
	if err != nil {
		return err
	}

	var existing bwitems
	err = json.Unmarshal(raw, &existing)
	if err != nil {
		return err
	}

	created, updated, removed, _, err := app.entriesToBitwarden(existing, entries)
	if err != nil {
		return err
	}

	for _, item := range created {
		fmt.Fprintf(os.Stderr, "creating in bitwarden: %s\n", item.Name)

		if app.dryRun {
			continue
		}

		err := app.createBitwardenItem(item)
		if err != nil {
			return err
		}
	}

	for _, item := range updated {
		fmt.Fprintf(os.Stderr, "updating in bitwarden: %s\n", item.Name)

		if app.dryRun {
			continue
		}

		err := app.updateBitwardenItem(item)
		if err != nil {
			return err
		}
	}

	if app.allowRemove {
		for _, item := range removed {
			fmt.Fprintf(os.Stderr, "removing from bitwarden: %s\n", item.Name)

			if app.dryRun {
				continue
			}

			err := app.deleteBitwardenItem(item)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
