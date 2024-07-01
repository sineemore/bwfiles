package main

import (
	"encoding/json"
	"path/filepath"
	"strconv"
)

type (
	bwfield struct {
		Name     string          `json:"name"`
		Value    string          `json:"value"`
		Type     int64           `json:"type"`
		LinkedID json.RawMessage `json:"linkedId"`
	}

	bwfields []bwfield

	bwitem struct {
		PasswordHistory json.RawMessage `json:"passwordHistory"`
		RevisionDate    json.RawMessage `json:"revisionDate"`
		CreationDate    json.RawMessage `json:"creationDate"`
		DeletedDate     json.RawMessage `json:"deletedDate"`
		Object          string          `json:"object"`
		ID              string          `json:"id"`
		OrganizationID  json.RawMessage `json:"organizationId"`
		FolderID        string          `json:"folderId"`
		Type            int64           `json:"type"`
		Reprompt        json.RawMessage `json:"reprompt"`
		Name            string          `json:"name"`
		Notes           string          `json:"notes"`
		Favorite        json.RawMessage `json:"favorite"`
		SecureNote      json.RawMessage `json:"secureNote"`
		CollectionIDs   json.RawMessage `json:"collectionIds"`
		Fields          bwfields        `json:"fields"`
	}

	bwitems []bwitem

	entry struct {
		name    string
		path    string
		mode    uint32
		content []byte
	}
)

const defaultTemplate string = `
{
	"passwordHistory": [],
	"revisionDate": null,
	"creationDate": null,
	"deletedDate": null,
	"organizationId": null,
	"collectionIds": null,
	"folderId": "",
	"type": 2,
	"name": "",
	"notes": "",
	"favorite": false,
	"fields": [],
	"login": null,
	"secureNote": {
		"type": 0
	},
	"card": null,
	"identity": null,
	"reprompt": 0
}`

func (item bwitem) same(content string, encoded bool, mode uint32) bool {
	encodedField, encodedOk := item.Fields.get("encoded")
	modeField, modeOk := item.Fields.get("mode")

	return item.Notes == content &&
		encodedOk && encodedField.Value == strconv.FormatBool(encoded) &&
		modeOk && modeField.Value == strconv.FormatUint(uint64(mode), 8)
}

func (items bwitems) get(name string) (bwitem, bool) {
	for _, item := range items {
		if item.Name == name {
			return item, true
		}
	}

	return bwitem{}, false
}

func (fields bwfields) get(name string) (bwfield, bool) {
	var (
		found bool
		field bwfield
	)

	for _, f := range fields {
		if f.Name == name {
			if found {
				return bwfield{}, false
			} else {
				found = true
				field = f
			}
		}
	}

	return field, found
}

func (app application) bitwardenwToEntries(items bwitems) ([]entry, error) {
	var entries []entry
	for _, item := range items {
		if item.Object != "item" || item.Type != 2 {
			continue
		}

		encodedField, ok := item.Fields.get("encoded")
		if !ok {
			continue
		}

		encoded, err := strconv.ParseBool(encodedField.Value)
		if err != nil {
			continue
		}

		modeField, ok := item.Fields.get("mode")
		if !ok {
			continue
		}

		mode, err := strconv.ParseUint(modeField.Value, 8, 32)
		if err != nil {
			continue
		}

		content, err := decode(item.Notes, encoded)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry{
			name:    item.Name,
			path:    filepath.Join(app.rootDirectory, item.Name),
			mode:    uint32(mode),
			content: content,
		})
	}

	return entries, nil
}

func (app application) newItem() (bwitem, error) {
	var item bwitem
	err := json.Unmarshal(app.config.BitwardenNewItemTemplate, &item)
	if err != nil {
		return bwitem{}, err
	}

	return item, nil
}

func (app application) entriesToBitwarden(existing bwitems, entries []entry) (bwitems, bwitems, bwitems, []entry, error) {
	var created, updated, removed bwitems
	var skipped []entry

	for _, e := range existing {
		found := false
		for _, entry := range entries {
			if e.Name == entry.name {
				found = true
				break
			}
		}

		if !found {
			removed = append(removed, e)
		}
	}

	for _, e := range entries {
		content, encoded, err := encode(e.content)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		item, exists := existing.get(e.name)
		if exists {
			if item.same(content, encoded, e.mode) {
				skipped = append(skipped, e)
				continue
			}
		} else {
			item, err = app.newItem()
			if err != nil {
				return nil, nil, nil, nil, err
			}
		}

		item.Name = e.name
		item.Notes = content
		item.FolderID = app.config.BitwardenFolderID
		item.Fields = []bwfield{
			{
				Name:  "encoded",
				Value: strconv.FormatBool(encoded),
				Type:  2,
			},
			{
				Name:  "mode",
				Value: strconv.FormatUint(uint64(e.mode), 8),
				Type:  0,
			},
		}

		if !exists {
			created = append(created, item)
		} else {
			updated = append(updated, item)
		}
	}

	return created, updated, removed, skipped, nil
}
