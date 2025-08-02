# deskctl
[![Go Report Card](https://goreportcard.com/badge/github.com/tzermias/deskctl)](https://goreportcard.com/report/github.com/tzermias/deskctl)

A tool to remote control Bluetooth-enabled standing desks.

It was created out of necessity, to have a fully-fledged tool to control my standing desk from the command line.

`deskcli` is based on findings from [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) where most UART commands were found.
`deskctl` just establishes a Bluetooth connection to the desk and issues respective commands. That's it.

> [!WARNING]
> `deskctl` is in WIP, with a limited set of features supported. Also, various bugs may surface as well.

## Prerequisites

* Bluetooth 4.0 adapter
* Standing desk with a [compatible controller](#supported-devices)

## Building 

```bash
make build
```
Upon successful build, the binary is located on `./bin/deskctl`

## Usage
Firstly, scan for supported devices
```bash
deskctl devices
```
For each supported device, MAC address, name and signal strength are shown.

Some example commands. Use the MAC address provided from `deskctl devices`

### Move up or down

It is equivalent of pushing once the up or down button on your desk.
```bash
deskctl -a <DEVICE_MAC_ADDRESS> up
deskctl -a <DEVICE_MAC_ADDRESS> down
```

### Move the desk to a specific height

Assuming that you just need to move the desk to an arbitrary height (e.g 107 cm), you can use the following command.
```bash 
deskctl -a <DEVICE_MAC_ADDRESS>  goto-height 107
```

### Go to a memory preset

Standing desks have usually up to 4 memory presets to store desk heights to. 
If you are already have configured a memory preset (e.g memory preset 1), you can move the desk to that preset.
```bash
deskctl -a <DEVICE_MAC_ADDRESS> goto-memory 1
```

## Supported devices

Currently desks with Jiecang controllers equipped with Lierda LSD4BT-E95ASTD001 BLE module are supported.
Controllers from other manufacturers may or may not work with this tool.

## Acknowledgements

Thanks to [phord/Jarvis](https://github.com/phord/Jarvis) and [pimp-my-desk/desk-control](https://gitlab.com/pimp-my-desk/desk-control) for reverse engineering the UART protocol used in Jiecang standing desks.
Also special thanks to [Cindy Xiao](https://cxiao.net) and their [post](https://cxiao.net/posts/2015-12-13-gatttool/) on how use `gatttool` to access data from BLE devices.
In fact, this post intrigued me in order to create a full fledged tool to control my standing desk.
