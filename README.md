# deej-winapp

**deej** is an open-source hardware volume mixer for Windows — physical sliders connected to an Arduino control your PC app volumes in real time.

This is a **fork** of [omriharel/deej](https://github.com/omriharel/deej), focused on fixing the Windows desktop client's stability issues.

## What's Different

The original Go desktop app has a recurring problem: after running for a while, the sliders stop responding while the tray icon still shows the app as running. On top of that, leftover `deej.exe` processes can pile up in the background and lock the COM port.

### Stability fixes

| Fix | What it does |
|---|---|
| **Auto-reconnect** | Serial drops → auto reconnects with exponential backoff (1s → 15s). No more silent death. |
| **Health monitor** | If no serial data for 5+ seconds, proactively rebuilds the connection. |
| **Single-instance lock** | Windows named mutex prevents multiple `deej.exe` instances from piling up. |
| **Serial format tolerance** | Accepts both `CRLF` and `LF` lines (original only took CRLF). |

### Multi slider-count support

The Go client **auto-detects** how many sliders your Arduino has — just look at the pipe-delimited serial data:

```
2 sliders → "423|512"       config slider_mapping: 0, 1
5 sliders → "423|512|0|128|900"     mapping: 0, 1, 2, 3, 4
8 sliders → "423|512|0|128|900|..." mapping: 0, 1, … 7
```

No code changes needed. Adjust your `config.yaml` to match.

## Quick Start

```powershell
git clone https://github.com/yueyaojade/deej-winapp.git
cd deej-winapp
scripts\build_windows.ps1
```

Or download the latest build from **Actions** tab → latest run → Artifacts.

Place `deej.exe` and `config.yaml` in the same folder. Edit `config.yaml` with your COM port and app mappings.

## Arduino Firmware

One sketch, any slider count. Flash `arduino/deej/deej.ino` to your Arduino board.

**To change the number of sliders**, edit just two lines at the top of the sketch:

```cpp
const int NUM_SLIDERS = 8;              // ← your count
const int analogInputs[] = {A0, A1, A2, A3, A4, A5, A6, A7};  // ← your pins
```

That's it. Supports 2 to 12 sliders out of the box.

## Configuration

See `config.yaml` in the repo root. Key fields:

```yaml
slider_mapping:
  0: master              # system master volume
  1: chrome.exe          # single app
  2: deej.current        # currently active window (Win only)
  3: deej.unmapped       # everything else
  # ...up to however many sliders you have

com_port: COM4
baud_rate: 9600
invert_sliders: false
noise_reduction: default
```

## Future planning

For the YAML configuration files, we may create related visual interfaces in the future to reduce the usage barrier of the project. (The original YAML will be retained to ensure compatibility with the main branch.)

## Build from Source (manual)

**Prerequisites:** Go 1.14+, mingw-w64, Git

```powershell
scripts\build_windows.ps1
```

Or use the batch script: `scripts\build_windows.bat`

## License

MIT — same as the [original project](https://github.com/omriharel/deej).

## Note

This project involves the participation of AI in the code writing process.
