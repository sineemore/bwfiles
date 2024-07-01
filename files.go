package main

import (
	"path/filepath"
	"strings"
)

type fileloc struct {
	path string
	name string
}

func (app application) selectFiles() ([]fileloc, error) {
	var includes []fileloc

	for _, p := range app.config.Patterns {
		glob := p.Glob
		if app.rootDirectory != "" {
			glob = filepath.Join(app.rootDirectory, glob)
		}

		if p.Include {
			matches, err := filepath.Glob(glob)
			if err != nil {
				return nil, err
			}

			for _, match := range matches {
				name := match
				fullpath := match
				if app.rootDirectory != "" {
					res, err := filepath.Rel(app.rootDirectory, match)
					if err != nil {
						return nil, err
					}

					if strings.HasPrefix(res, "..") || strings.HasPrefix(res, "/") {
						continue
					}

					name = "/" + res
				}

				includes = append(includes, fileloc{
					path: fullpath,
					name: name,
				})
			}
		} else {
			for i := len(includes) - 1; i >= 0; i-- {
				m, err := filepath.Match(glob, includes[i].path)
				if err != nil {
					return nil, err
				}

				if m {
					includes = append(includes[:i], includes[i+1:]...)
				}
			}
		}
	}

	return includes, nil
}
