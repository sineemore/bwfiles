bwfiles
=======

Sync local files as secure notes to bitwarden using bw cli.

- all bitwarden operations via `bw` (see [exec.go](exec.go))
- `bwfiles` never gets full list of your credentials
- no external dependencies/packages

Sample config `.bwfilesrc`:

```json
{
    "bitwarden_folder_id": "1bf5d349-063e-46e7-92f6-94d93c7048b7",
    "patterns": [
        "/home/user/.ssh/*",
        "!/home/user/.ssh/*known_hosts*"
    ]
}
```