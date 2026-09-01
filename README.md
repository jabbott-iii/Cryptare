## Features:

- **File Encryption & Decryption**
  - Encrypt files with AES-256-GCM
  - Decrypt previously encrypted files
  - Optional custom output paths
  - Password-based protection for secure workflows

- **Compression & Decompression**
  - Compress files with gzip
  - Decompress gzip archives
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

- cryptare encrypt — encrypt a file
- cryptare decrypt — decrypt a file
- cryptare compress — compress a file with gzip
- cryptare decompress — decompress a gzip file
- cryptare keys — manage encryption keys

### encrypt

- cryptare encrypt [file] — encrypt a file with AES-256-GCM
- cryptare encrypt [file] --output [path] — write to a custom output file
- cryptare encrypt [file] --password [value] — provide the encryption password non-interactively

Examples:
- cryptare encrypt ./secret.txt
- cryptare encrypt ./secret.txt --output ./secret.txt.enc
- cryptare encrypt ./secret.txt --password "correct horse battery staple"

### decrypt

- cryptare decrypt [file] — decrypt an AES-256-GCM encrypted file
- cryptare decrypt [file] --output [path] — write to a custom output file
- cryptare decrypt [file] --password [value] — provide the decryption password non-interactively

Examples:
- cryptare decrypt ./secret.txt.enc
- cryptare decrypt ./secret.txt.enc --output ./secret.txt
- cryptare decrypt ./secret.txt.enc --password "correct horse battery staple"

### compress

- cryptare compress [file] — compress a file with gzip
- cryptare compress [file] --output [path] — write to a custom output file
- cryptare compress [file] --level [1-9] — set gzip compression level

Examples:
- cryptare compress ./artifact.bin
- cryptare compress ./artifact.bin --output ./artifact.bin.gz
- cryptare compress ./artifact.bin --level 9

### decompress

- cryptare decompress [file] — decompress a gzip file
- cryptare decompress [file] --output [path] — write to a custom output file

Examples:
- cryptare decompress ./artifact.bin.gz
- cryptare decompress ./artifact.bin.gz --output ./artifact.bin

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
