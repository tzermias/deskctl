# deskctl

A CLI tool to manage Bluetooth standing desks.
Currently limited to Jiecang standing desks


## Prerequisites

A Bluetooth 4.0 adapter

## Installation
```bash
go get github.com/tzermias/deskctl@latest
```

## Usage
Find supported devices
```bash
deskctl devices
```

Go to preset from Memory 1
```bash
deskctl goto-memory -a <DEVICE_MAC_ADDRESS> --memory 1
```
