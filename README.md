# deej-winapp

**deej** is an open-source hardware volume mixer for Windows — physical sliders connected to an Arduino control your PC app volumes in real time.

This is a **fork** of [omriharel/deej](https://github.com/omriharel/deej), focused on fixing the Windows desktop client's stability issues.

## What's Different

The original Go desktop app has a recurring problem: after running for a while, the sliders stop responding while the tray icon still shows the app as running. On top of that, leftover `deej.exe` processes can pile up in the background and lock the COM port.

| Fix | File | What it does |
|---|---|---|
| **Auto-reconnect on disconnect** | `pkg/deej/serial.go` | When the serial connection drops (USB unplug, Arduino reset, COM port glitch), the app now automatically reconnects with exponential backoff (1s → 15s). No more silent death. |
| **Connection health monitor** | `pkg/deej/serial.go` | If no serial data arrives for 5+ seconds, the connection is proactively closed and rebuilt. |
| **Single-instance lock** | `pkg/deej/cmd/single_instance_windows.go` | Windows named mutex prevents multiple `deej.exe` instances from running simultaneously. Second instance exits immediately. |
| **Serial format tolerance** | `pkg/deej/serial.go` | Accepts both `CRLF` and `LF`-terminated lines (original only accepted CRLF). |
| **8-slider Arduino sketch** | `arduino/deej-8-sliders/` | 8-channel variant with 50ms loop interval (vs original 10ms) for better Windows USB-Serial stability. |
| **Build scripts** | `scripts/` | PowerShell and batch scripts for easy Windows compilation. |

## Build from Source

**Prerequisites:**
- [Go](https://go.dev/dl/) 1.14+
- [mingw-w64](http://mingw-w64.org/) (CGo requirement for systray)
- Git

```powershell
git clone https://github.com/yueyaojade/deej-winapp.git
cd deej-winapp
scripts\build_windows.ps1
```

Output: `deej.exe` — place alongside `config.yaml` from this repo.

## Arduino Firmware

Flash `arduino/deej-8-sliders/deej-8-sliders.ino` to your Arduino Nano (or adjust pins for your board). The 5-slider original is also available at `arduino/deej-5-sliders-vanilla/`.

### Schematic

A standard deej wiring setup — see the [original project](https://github.com/omriharel/deej) for detailed hardware instructions.

## Configuration

Place `config.yaml` next to `deej.exe`. Example for 8 sliders:

```yaml
slider_mapping:
  0: master
  1: chrome.exe
  2: spotify.exe
  3: discord.exe
  4: deej.unmapped
  5: deej.current
  6: obs64.exe
  7: mic

com_port: COM4
baud_rate: 9600
noise_reduction: default
```

## License

MIT — same as the [original project](https://github.com/omriharel/deej).
