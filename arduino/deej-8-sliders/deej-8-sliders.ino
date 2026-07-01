/*
  deej - 8-slider version

  Modified for 8-channel operation with reduced serial frequency
  for improved Windows compatibility.

  Original: omriharel/deej (5-slider vanilla)
*/

const int NUM_SLIDERS = 8;
const int analogInputs[NUM_SLIDERS] = {A0, A1, A2, A3, A4, A5, A6, A7};

int analogSliderValues[NUM_SLIDERS];

void setup() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  Serial.begin(9600);

  // Small delay after serial init to let the host catch up
  delay(500);
}

void loop() {
  updateSliderValues();
  sendSliderValues();

  // 50ms interval = 20 updates/second.
  // This is well within human reaction speed and reduces serial load
  // compared to the original 10ms, which improves Windows stability.
  delay(50);
}

void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
     analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void sendSliderValues() {
  String builtString = String("");

  for (int i = 0; i < NUM_SLIDERS; i++) {
    builtString += String((int)analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }

  Serial.println(builtString);
}
