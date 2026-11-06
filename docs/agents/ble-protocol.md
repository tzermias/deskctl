# BLE Protocol & `pkg/jiecang`

The desks use Jiecang controllers with Lierda LSD4BT-E95ASTD001 BLE modules.

## GATT

| UUID | Purpose |
|------|---------|
| `0xFE60` | Jiecang service |
| `0xFE61` | DataIn: write commands |
| `0xFE62` | DataOut: notifications/responses |

## Message format (Jiecang UART protocol)

- Commands sent: `[0xf1, 0xf1, command, length, data..., checksum, 0x7e]`
- Responses received: same layout, but they start with `0xf2, 0xf2`.
- Checksum = (command + length + sum(data)) mod 256. See
  [Framing](#framing).
- Validate responses with `isValidData` in `pkg/jiecang/common.go`. Don't
  write a separate validator.

## Thread safety

`Jiecang` guards shared state (`currentHeight`, `presets`, height limits)
with `mu sync.RWMutex`. Notification callbacks run on a different
goroutine, so hold `j.mu` (RLock to read, Lock to write) whenever you
access these fields.

## Framing

The BLE module (JCP35NBLT dongle / Lierda module) is a bridge to the
controller's handset UART on the RJ-12 port, at 9600 8N1 [JV]. Each BLE
write is one UART frame and each notification carries one or more frames.

- Address `F1F1` = handset (us), `F2F2` = controller [JV].
- Checksum = 8-bit sum of command, length and data, i.e. mod 256. The
  Jarvis README says `% 0xFF`, but its code overflows an `unsigned char`,
  so it really is mod 256 [JV]. That's what `isValidData` checks.
- Command and response codes share one number space: `0E`, `19` and `1D`
  set a value when sent with one parameter and report it when received.

## Commands

| Code | Name in `commands` | Payload | Implemented | Source |
|------|--------------------|---------|-------------|--------|
| `01` | `up` | none | yes | [JV], this codebase |
| `02` | `down` | none | yes | [JV], this codebase |
| `03` / `04` / `25` / `26` | `save_memory1`-`3` (PROGMEM 1-4) | none | 1-3 | [JV], this codebase |
| `05` / `06` / `27` / `28` | `goto_memory1`-`3` (MOVE 1-4) | none | 1-3 | [JV], this codebase |
| `07` | `fetch_height` (really SETTINGS) | none | yes | [JV], this codebase |
| `0C` | `fetch_height_range` | none | yes | [JV], this codebase |
| `0E` | set units | 1 byte | no | [JV] |
| `19` | set memory mode (`00` one-touch, `01` constant) | 1 byte | no | [JV] |
| `1B` | go to height | `mm/256 mm%256` | yes (`GoToHeight`) | [RE] |
| `1D` | set anti-collision sensitivity | 1 byte | no | [JV] |
| `20` | fetch high/low limit | none | no | [RE], [JV] |
| `21` / `22` | set max / min limit | | no | [JV] |
| `23` | clear limits | | no | [JV] |
| `2B` | `stop` | none | yes | [RE] |
| `40` | reset | | no | [JV] |
| `91` | calibrate | | no | [JV] |
| `A0` | patch (settings?) | | no | [RE] |
| `A2` | `fetch_stand_time` | none | yes | [RE] |
| `AA` | `fetch_all_time` | none | yes | this codebase |
| `AD` | set memory height | `slot(1-3) mm_hi mm_lo` | no | [RE] |

Notes:

- **Stand time is ambiguous.** The [RE] command table says `A2 00`, but
  the decompiled app has `cmdFetchStandTime = F1F1A60100A77E` (`A6`, one
  byte `00`). We send `A2`. Neither stand-time nor all-time replies are
  decoded.
- **PROGMEM (save memory) is never acknowledged** [JV]. The cached preset
  only changes after the next `FetchHeight`.
- Every entry in `commands` must pass `TestCommandChecksums`.

## Responses

| Code | Len | Meaning | Handled by `characteristicReceiver` | Source |
|------|-----|---------|------------------------------|--------|
| `01` | 3 | height `{P0,P1}`, `P2` unknown (`07`/`0F`) | `currentHeight` via `readHeight` | [JV] |
| `07` | 4 | `{P0,P1}` upper, `{P2,P3}` lower limit in mm | `HighestHeight`/`LowestHeight` via `readHeightRange` | [JV] |
| `0E` | 1 | units (mm / inch) | printed only | [JV] |
| `17` | | unknown | ignored | [JV] |
| `19` | 1 | memory mode (`00` one-touch, `01` constant) | `MemoryConstantTouchMode` (buggy, see below) | [JV] |
| `1B` | | ack of go to height | ignored | this codebase |
| `1D` | 1 | anti-collision sensitivity | `AntiCollisionSensitivity` (buggy, see below) | [JV] |
| `20` | 1 | limits set (`00`/`01`/`10`/`11`) | logged | [JV] |
| `21` / `22` | | max / min limit | logged | [JV] |
| `23` | | limits cleared / limit stop | logged | [JV] |
| `25`-`28` | 2 | memory preset 1-4 in mm | `presets["memoryN"]` via `readMemoryPreset` | [JV] |
| `92` | | moving to preset (not sent on RJ-12) | logged | [JV] |

### `07` is the SETTINGS request

`fetch_height` (`07`) does not just return the height. Over RJ-12 the desk
replies with the presets `25 26 27 28` and then the height `01`. Over RJ-45
it also sends `0E 19 17 1D` [JV]. `FetchHeight` therefore fills both
`presets` and `currentHeight`.

### Height units

Heights are in mm, **or tenths of an inch when the desk is set to inches**
(range 240..530) [JV]. `readHeight`, `readHeightRange` and
`readMemoryPreset` always assume mm and convert to whole cm.

### Known bugs (pinned by `TODO` tests, not yet fixed)

- The mm to cm conversion uses integer division, so it truncates instead
  of rounding (827 mm becomes 82 cm).
- Inch units are ignored.
- `19` and `1D` handling reads byte 3 (the length, always `01`) instead of
  byte 4, so memory mode is always set to constant touch and is never
  cleared, and sensitivity always reads as 1.
- Frames are split on every `0x7E` byte, so a frame whose payload contains
  `0x7E` (e.g. height 894 mm = `03 7E`) is dropped.
- `GoToHeight` returns `nil` on cancel, although its doc says `ctx.Err()`.
- `GoToMemory` doesn't send `stop` on cancel.

## Test seam

`Jiecang` stores the device and characteristics behind two unexported
interfaces in `jiecang.go`:

- `characteristic`: `Read`, `WriteWithoutResponse`, `EnableNotifications`
  (the subset of `bluetooth.DeviceCharacteristic` we use)
- `disconnecter`: `Disconnect`

`var _ characteristic = (*bluetooth.DeviceCharacteristic)(nil)` keeps the
real type compatible. Use a pointer: on Linux `EnableNotifications` has a
pointer receiver. (On macOS tinygo's `DeviceCharacteristic` has no `Read`,
so the package doesn't build there today.)

`Init` only connects and discovers the service and characteristics, then
calls `newJiecang(dev, in, out)`, which does everything else: enable
notifications, initial read, presets map, and the four `Fetch*` calls.
Tests call `newJiecang` with a fake desk; see [testing.md](testing.md).
Movement loops poll every `pollInterval` (200ms by default).

## References

- [RE] jiecang-reverse-engineering, command table and decompiled BLE dongle
  app strings:
  https://gitlab.com/pimp-my-desk/desk-control/jiecang-reverse-engineering#source
- [JV] phord/Jarvis, README sections "UART protocol", "Known Commands" and
  "Known Responses", and `src/ProtocolFully.cpp`:
  https://github.com/phord/Jarvis
