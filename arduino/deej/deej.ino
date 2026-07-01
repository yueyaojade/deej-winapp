/*
 * deej — Hardware volume mixer for Windows
 *
 * Reads slider pots and sends values over USB serial.
 * The connected Go desktop client adjusts app volumes accordingly.
 *
 * ─── How to configure ───
 *   NUM_SLIDERS: how many sliders you have (2 to 12)
 *   analogInputs[]: which analog pins they're connected to
 *   SERIAL_INTERVAL_MS: milliseconds between each send (50 = 20 times/sec)
 *
 *   That's it. Everything else is automatic.
 *
 * ─── Wiring ───
 *   Each slider pot: outer legs → 5V and GND, middle leg → analog pin
 *   No external components needed (Arduino's internal pull-ups are unused;
 *   the pots provide their own voltage divider).
 *
 * ─── Serial output format ───
 *   423|512|1023|0|128|...
 *   Pipe-delimited integers, one per slider, range 0–1023.
 *   Terminated with CRLF.
 *
 * License: MIT
 */

// ═══════════════════════════════════════════════════════════════════════
//  YOU ONLY NEED TO CHANGE THESE TWO LINES
// ═══════════════════════════════════════════════════════════════════════

const int NUM_SLIDERS = 8;         // ← change this to match your build
const int analogInputs[] = {        // ← list the analog pins you're using
  A0, A1, A2, A3, A4, A5, A6, A7  // 8 sliders → 8 pins
};

// ═══════════════════════════════════════════════════════════════════════
//  OPTIONAL: tweak timing
// ═══════════════════════════════════════════════════════════════════════

// Sending interval (milliseconds):
//   50ms → 20 updates/sec (recommended for Windows stability)
//   10ms → 100 updates/sec (original, more USB serial pressure)
//   100ms → 10 updates/sec (lighter, fine for most use)
const int SERIAL_INTERVAL_MS = 50;

// ═══════════════════════════════════════════════════════════════════════
//  NO CHANGES NEEDED BELOW THIS LINE
// ═══════════════════════════════════════════════════════════════════════

int analogSliderValues[NUM_SLIDERS];

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  Serial.begin(9600);

  // Wait for the USB-serial connection to stabilize before sending data
  // (especially important on Windows with CH340/FT232 adapters)
  delay(500);
}

void loop() {
  readSliderValues();
  sendSliderValues();
  delay(SERIAL_INTERVAL_MS);
}

void readSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void sendSliderValues() {
  String line = "";

  for (int i = 0; i < NUM_SLIDERS; i++) {
    line += String(analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      line += "|";
    }
  }

  Serial.println(line);
}
