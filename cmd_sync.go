package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

		entries = append(entries, entry{
			name:    file.name,
			path:    file.path,
			mode:    uint32(info.Mode()),
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

	created, updated, removed, skipped, err := app.entriesToBitwarden(existing, entries)
	if err != nil {
		return err
	}

	for _, e := range skipped {
		fmt.Printf("skipping %s\n", e.path)
	}

	for _, item := range created {
		fmt.Printf("creating %s\n", item.Name)

		if app.dryRun {
			continue
		}

		err := app.createBitwardenItem(item)
		if err != nil {
			return err
		}
	}

	for _, item := range updated {
		fmt.Printf("updating %s\n", item.Name)

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
			fmt.Printf("removing %s\n", item.Name)

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
