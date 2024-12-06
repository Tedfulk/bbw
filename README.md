# BBW (Better Bitwarden)

A terminal user interface wrapper for the Bitwarden CLI that provides an enhanced, interactive experience for managing your Bitwarden vault.

## Features

- 🔍 Interactive search with fuzzy matching
- 🔐 Password and passphrase generation
- 📋 Quick clipboard copying for credentials
- 🔄 Vault syncing and status checking
- 📱 User-friendly terminal interface
- 🔑 Secure session management
- 📊 Detailed item metadata viewing

## Prerequisites

- [Bitwarden CLI](https://bitwarden.com/help/cli/) installed and available in your PATH
- Go 1.21 or later

## Installation

```sh
go install github.com/Tedfulk/bbw@latest
```

## Configuration

BBW stores its configuration in `~/.config/bbw/config.yaml`. The configuration file will be created automatically on first run.

## Usage

### Basic Commands

- `s` - Show vault status
- `g` - Generate password/passphrase
- `h` - Show help
- `u` - Update CLI and sync vault
- `q` - Quit

### Search

Simply start typing to search your vault. Use arrow keys to navigate results and Enter to select.

### Password Generation

Three types of password generation are available:

1. Password with special characters (18 characters, minimum 2 special chars, 2 numbers)
2. Password without special characters (18 characters, minimum 2 numbers)
3. Passphrase (5 words with a number)

### Item Management

When viewing an item, you can:

- Copy username
- Copy password
- Copy notes (if available)
- Copy URL (if available)
- View detailed metadata including creation date, last modified, and password history
