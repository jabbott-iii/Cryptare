## Features:

- **File Encryption & Decryption**
  - Encrypt files with AES-256-GCM
  - Encrypt directories into a single encrypted archive
  - Decrypt previously encrypted files
  - Decrypt encrypted directory archives back into their original tree
  - Optional custom output paths
  - Password-based protection for secure workflows

- **Compression & Decompression**
  - Compress files with gzip
  - Compress directories as tar.gz archives
  - Decompress gzip files and tar.gz archives
  - Optional compression levels
  - Automatic output name derivation

- **Key Management**
  - Generate and store encryption keys
  - List stored keys
  - Export keys to encrypted files
  - Import keys from encrypted files

- **Interactive TUI**
  - Terminal user interface for file and key management
  - Launches automatically when run without subcommands

## Core CLI capabilities

Cryptare is organized into focused command groups:

- cryptare encrypt — encrypt a file or directory
- cryptare decrypt — decrypt a file or encrypted directory archive
- cryptare compress — compress a file or directory
- cryptare decompress — decompress a gzip file or tar.gz archive
- cryptare keys — manage encryption keys

### encrypt

- cryptare encrypt [path] — encrypt a file with AES-256-GCM or package a directory into a single encrypted archive
- cryptare encrypt [path] --output [path] — write to a custom output file
- cryptare encrypt [path] --password [value] — provide the encryption password non-interactively

Examples:
- cryptare encrypt ./secret.txt
- cryptare encrypt ./secret.txt --output ./secret.txt.enc
- cryptare encrypt ./project-dir --output ./project-dir-backup.enc
- cryptare encrypt ./secret.txt --password "correct horse battery staple"

### decrypt

- cryptare decrypt [path] — decrypt an AES-256-GCM encrypted file or restore an encrypted directory archive
- cryptare decrypt [path] --output [path] — write to a custom output file or restore into a target directory
- cryptare decrypt [path] --password [value] — provide the decryption password non-interactively

Examples:
- cryptare decrypt ./secret.txt.enc
- cryptare decrypt ./secret.txt.enc --output ./secret.txt
- cryptare decrypt ./project-dir-backup.enc --output ./restored-project-dir
- cryptare decrypt ./secret.txt.enc --password "correct horse battery staple"

### compress

- cryptare compress [path] — compress a file with gzip or a directory as tar.gz
- cryptare compress [path] --output [path] — write to a custom output file or archive
- cryptare compress [path] --level [1-9] — set gzip compression level

Examples:
- cryptare compress ./artifact.bin
- cryptare compress ./artifact.bin --output ./artifact.bin.gz
- cryptare compress ./build --output ./build-backup.tar.gz
- cryptare compress ./artifact.bin --level 9

### decompress

- cryptare decompress [archive] — decompress a gzip file or extract a tar.gz archive
- cryptare decompress [archive] --output [path] — write to a custom output file or extract to a directory

Examples:
- cryptare decompress ./artifact.bin.gz
- cryptare decompress ./artifact.bin.gz --output ./artifact.bin
- cryptare decompress ./build-backup.tar.gz --output ./restored-build

### keys

- cryptare keys list — list stored encryption keys
- cryptare keys generate — generate and store a new random encryption key
- cryptare keys export [key-id] — export an encrypted key to a file
- cryptare keys import [file] — import an encrypted key from a file

Examples:
- cryptare keys list
- cryptare keys generate --password "master password"
- cryptare keys export key-123 --output key-123.ckey
- cryptare keys import ./key-123.ckey

### Interactive TUI

- Running cryptare with no subcommand launches the terminal UI dashboard for interactive file and key management.

## Install:

Download the appropriate binary for your platform below and make it executable:

Linux:
```
chmod +x cryptare-linux-amd64
```
```
sudo mv cryptare-linux-amd64 /usr/local/bin/cryptare
```
 or
```
chmod +x cryptare-linux-arm64
```
```
sudo mv cryptare-linux-arm64 /usr/local/bin/cryptare
```
macOS:
```
chmod +x cryptare-macos-arm64
```
```
sudo mv cryptare-macos-arm64 /usr/local/bin/cryptare
```
  or
```
chmod +x cryptare-macos-amd64
```
```
sudo mv cryptare-macos-amd64 /usr/local/bin/cryptare
```
Windows:
```
Download cryptare-windows-amd64.exe and add it to your PATH as cryptare.
```

## Docker

### Build
```bash
docker build -t cryptare:latest .
```

### Run (interactive)
```bash
docker run --rm -it cryptare:latest
```

### Persist data
```bash
mkdir -p ~/.cryptare
docker run --rm -it \
  -v ~/.cryptare:/app/data \
  -e CRYPTARE_DB_PATH=/app/data/cryptare.db \
  cryptare:latest
```

### CLI usage
```bash
docker run --rm -it cryptare:latest --help
docker run --rm -it -v ~/.cryptare:/app/data cryptare:latest keys list
```
