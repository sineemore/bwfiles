package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type (
	commandsConfig struct {
		GetItems   string `json:"get_items"`
		CreateItem string `json:"create_item"`
		UpdateItem string `json:"update_item"`
		DeleteItem string `json:"delete_item"`
	}

	pattern struct {
		Glob    string
		Include bool
	}

	config struct {
		Commands                 commandsConfig  `json:"commands"`
		BitwardenFolderID        string          `json:"bitwarden_folder_id"`
		BitwardenNewItemTemplate json.RawMessage `json:"bitwarden_new_item_template"`
		Patterns                 []pattern       `json:"patterns"`
	}
)

func (p *pattern) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	if s[0] == '!' {
		p.Include = false
		p.Glob = s[1:]
	} else {
		p.Include = true
		p.Glob = s
	}

	return nil
}

func (p *pattern) MarshalJSON() ([]byte, error) {
	if !p.Include {
		return json.Marshal("!" + p.Glob)
	}

	return json.Marshal(p.Glob)
}

func loadConfig(configPath string) (config, error) {
	configLocations := []string{
		filepath.Join(os.Getenv("XDG_CONFIG_HOME"), "bwfiles/config.json"),
		filepath.Join(os.Getenv("HOME"), ".bwfilesrc"),
	}

	if os.Getenv("BWFILES_CONFIG") != "" {
		configLocations = append([]string{os.Getenv("BWFILES_CONFIG")}, configLocations...)
	}

	if configPath != "" {
		configLocations = append([]string{configPath}, configLocations...)
	}

	var cfg config
	cfg.Commands = defaultCommands
	cfg.BitwardenNewItemTemplate = []byte(defaultTemplate)
	cfg.Patterns = []pattern{}

	for _, loc := range configLocations {
		f, err := os.Open(loc)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return config{}, err
		}
		defer f.Close()

		err = json.NewDecoder(f).Decode(&cfg)
		if err != nil {
			return config{}, err
		}

		return cfg, nil
	}

	return cfg, nil
}
