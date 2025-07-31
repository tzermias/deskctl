# deskctl

A CLI tool to manage Bluetooth standing desks.
Currently limited to Jiecang standing desks

Tool is based on findings from [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) where most UART commands were found.
`deskctl` just establishes a Bluetooth connection to the desk and issues respective commands. That's it.

The tool is under development and only a small subset of features (moving the desk to memory presets) is supported at the moment.

## Prerequisites

* Bluetooth 4.0 adapter
* Jiecang compatible standing desk (of course :P)

## Building 
```bash
go build -o deskctl
```

## Usage
Find supported devices
```bash
deskctl devices
```

Go to memory preset number 1 (using the MAC address of the desk as `<DEVICE_MAC_ADDRESS>`)
```bash
deskctl goto-memory -a <DEVICE_MAC_ADDRESS> 1
```

## Acknowledgements

Thanks to [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) for reverse engineering the UART protocol used in Jiecang standing desks.
Also special thanks to [Cindy Xiao](https://cxiao.net) and their [post](https://cxiao.net/posts/2015-12-13-gatttool/) on how use `gatttool` to access data from BLE devices.
In fact, this post intrigued me in order to create a full fledged tool to control my standing desk.
