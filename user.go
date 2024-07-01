package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

func getIdentity(identity string) (uint32, uint32, error) {
	if identity == "" {
		return 0, 0, errors.New("identity not provided")
	}

	parts := strings.SplitN(identity, ":", 2)

	u, err := user.Lookup(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("failed to lookup user %s: %w", parts[0], err)
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse uid %s: %w", u.Uid, err)
	}

	var gid uint64

	if len(parts) == 1 {
		gid, err = strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse gid %s: %w", u.Gid, err)
		}
	} else {
		g, err := user.LookupGroup(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("failed to lookup group %s: %w", parts[1], err)
		}

		gid, err = strconv.ParseUint(g.Gid, 10, 32)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse gid %s: %w", g.Gid, err)
		}
	}

	return uint32(uid), uint32(gid), nil
}

func checkRunningAsRoot(identity string, allowRunningAsRoot bool) (bool, bool, uint32, uint32, error) {
	current, err := user.Current()
	if err != nil {
		return false, false, 0, 0, fmt.Errorf("failed to get current user: %w", err)
	}

	if current.Uid != "0" || allowRunningAsRoot {
		// Allow running directly as current user (even root)
		return current.Uid == "0", false, 0, 0, nil
	}

	// Running as root, need to drop privileges

	if identity == "" {
		if os.Getenv("SUDO_USER") != "" {
			identity = os.Getenv("SUDO_USER")
		} else {
			return false, false, 0, 0, errors.New("identity not provided")
		}
	}

	uid, gid, err := getIdentity(identity)
	if err != nil {
		return false, false, 0, 0, fmt.Errorf("failed to get identity: %w", err)
	}

	return true, true, uid, gid, nil
}
