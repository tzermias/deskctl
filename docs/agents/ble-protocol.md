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
- Checksum = (command + length + sum(data)) mod 256.
- Validate responses with `isValidData` in `pkg/jiecang/common.go`. Don't
  write a separate validator.

## Thread safety

`Jiecang` guards shared state (`currentHeight`, `presets`, height limits)
with `mu sync.RWMutex`. Notification callbacks run on a different
goroutine, so hold `j.mu` (RLock to read, Lock to write) whenever you
access these fields.
