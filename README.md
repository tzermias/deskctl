# deskctl

A CLI tool to manage Bluetooth standing desks.
Currently limited to Jiecang standing desks

Tool is based on findings from [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) where most UART commands were found.
`deskcli` just establishes a Bluetooth connection to the desk and issues respective commands. That's it.

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

## Acknowledgements

Thanks to [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) for reverse engineering the UART protocol used in Jiecang standing desks.
Also special thanks to [Cindy Xiao](https://cxiao.net) and their [post](https://cxiao.net/posts/2015-12-13-gatttool/) on how use `gatttool` to access data from BLE devices.
In fact, this post intrigued me in order to create a full fledged tool to control my standing desk.
