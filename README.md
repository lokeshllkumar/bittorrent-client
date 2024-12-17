# bittorent-client

A light-weight BitTorrent client written in Go, providing a robust peer-to-peer file sharing solution.

## Features

- Torrent Parsing
    - Reads .torrent files to extract metadata.
- Peer Communication
    - Establishes communication with peers using TCP.
- Bitfield Handling
    - Processes peer bitfields to track peer availability.
- Timeout Management
    - Configurable timeouts for read/write operations to handle network delays or unresponsive peers.

## Setup and Installation

- Prerequisites
    - Go 1.20+
- Clone the Repository
```bash
git clone https://github.com/lokeshllkumar/bittorrent-client
cd bittorrent-client
```
- Installing dependencies
```bash
go mod tidy
```
- Build the Project
```bash
go build .
```

## Usage

Execute the program with a ```.torrent``` file.

```bash
./bittorent-client <path-to-torrent-file> <path-to-output-file>
```

The client will parse the torrent file, retrieve information about peers, connect to peers and exchange handshake messages, and download the file in pieces from multiple peers.