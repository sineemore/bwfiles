package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"syscall"
)

var defaultCommands = commandsConfig{
	GetItems:   "bw --nointeraction list items --folderid $BW_FOLDER_ID",
	CreateItem: "bw encode | bw --nointeraction create item >/dev/null",
	UpdateItem: "bw encode | bw --nointeraction edit item $BW_ITEM_ID >/dev/null",
	DeleteItem: "bw --nointeraction delete item $BW_ITEM_ID </dev/null",
}

func (app application) execCommand(command string, item *bwitem) ([]byte, error) {
	cmd := exec.Command("sh", "-c", command)
	cmd.Env = append(os.Environ(), "BW_FOLDER_ID="+app.config.BitwardenFolderID)

	if item != nil {
		cmd.Env = append(cmd.Env, "BW_ITEM_ID="+item.ID)

		b, err := json.Marshal(item)
		if err != nil {
			return nil, err
		}

		cmd.Stdin = bytes.NewReader(b)
	}

	if app.dropPrivileges {
		cmd.Dir = "/"
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Credential: &syscall.Credential{
				Uid:    app.uid,
				Gid:    app.gid,
				Groups: []uint32{},
			},
			Setsid: true,
		}
	}

	return cmd.Output()
}

func (app application) getBitwardenItems() ([]byte, error) {
	return app.execCommand(app.config.Commands.GetItems, nil)
}

func (app application) createBitwardenItem(item bwitem) error {
	_, err := app.execCommand(app.config.Commands.CreateItem, &item)
	return err
}

func (app application) updateBitwardenItem(item bwitem) error {
	_, err := app.execCommand(app.config.Commands.UpdateItem, &item)
	return err
}

func (app application) deleteBitwardenItem(item bwitem) error {
	_, err := app.execCommand(app.config.Commands.DeleteItem, &item)
	return err
}
