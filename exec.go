package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
)

var defaultCommands = commandsConfig{
	GetItems:   "bw list items --folderid $BW_FOLDER_ID",
	CreateItem: "bw encode | bw create item >/dev/null",
	UpdateItem: "bw encode | bw edit item $BW_ITEM_ID >/dev/null",
	DeleteItem: "bw delete item $BW_ITEM_ID </dev/null",
}

func (app application) getBitwardenItems() ([]byte, error) {
	cmd := exec.Command("sh", "-c", app.config.Commands.GetItems)
	cmd.Env = append(
		os.Environ(),
		"BW_FOLDER_ID="+app.config.BitwardenFolderID,
	)
	return cmd.Output()
}

func (app application) execCommand(command string, item bwitem) error {
	b, err := json.Marshal(item)
	if err != nil {
		return err
	}

	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(
		os.Environ(),
		"BW_ITEM_ID="+item.ID,
		"BW_FOLDER_ID="+app.config.BitwardenFolderID,
	)
	cmd.Stdin = bytes.NewReader(b)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (app application) createBitwardenItem(item bwitem) error {
	return app.execCommand(app.config.Commands.CreateItem, item)
}

func (app application) updateBitwardenItem(item bwitem) error {
	return app.execCommand(app.config.Commands.UpdateItem, item)
}

func (app application) deleteBitwardenItem(item bwitem) error {
	return app.execCommand(app.config.Commands.DeleteItem, item)
}
