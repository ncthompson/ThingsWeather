/*
 * Author: NC Thompson
 * Date: 2017-02-15
 *
 */
#include <rn2xx3.h>
#include <SHT1x.h>
#include "LowPower.h"

//create an instance of the rn2483 library, using the given Serial port
rn2xx3 myLora(Serial1);
#define dataPin  SDA
#define clockPin SCL
#define PIN_SERIAL1_TX PIN1

SHT1x sht1x(dataPin, clockPin);

uint8_t count = 0;

uint8_t txBuffer[7];
int batVoltPin = A5;    // select the input pin for the battery voltage

// the setup routine runs once when you press reset:
void setup()
{
  //output LED pin
  pinMode(13, OUTPUT);
  led_on();

  // Open serial communications and wait for port to open:
  Serial.begin(57600); //serial port to computer

  Serial1.begin(57600); //serial port to radio
  //wakeUP_RN2483();
  // make sure usb serial connection is available,
  // or after 10s go on anyway for 'headless' use of the
  // node.
  while ((!Serial) && (millis() < 10000));

  Serial.println("Startup");

  initialize_radio();

  //transmit a startup message
  myLora.tx("TTN Mapper on TTN Uno node");
  Serial.end();
  led_off();
  delay(2000);
}

void initialize_radio()
{
  delay(100); //wait for the RN2xx3's startup message
  Serial1.flush();

  //print out the HWEUI so that we can register it via ttnctl
  String hweui = myLora.hweui();
  while(hweui.length() != 16)
  {
    Serial.println("Communication with RN2xx3 unsuccessful. Power cycle the TTN UNO board.");
    delay(10000);
    hweui = myLora.hweui();
  }
  Serial.println("When using OTAA, register this DevEUI: ");
  Serial.println(hweui);
  Serial.println("RN2xx3 firmware version:");
  Serial.println(myLora.sysver());

  //configure your keys and join the network
  Serial.println("Trying to join TTN");
  bool join_result = false;

  myLora.setFrequencyPlan(TTN_EU);

  //ABP: initABP(String addr, String AppSKey, String NwkSKey);
  join_result = myLora.initABP("XXXXXXXX", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX");

  while(!join_result)
  {
    Serial.println("Unable to join. Are your keys correct, and do you have TTN coverage?");
    delay(60000); //delay a minute before retry
    join_result = myLora.init();
  }
  Serial.println("Successfully joined TTN");
  okSeq();

}
void okSeq () {
    led_on();
    delay(100);
    led_off();
    delay(100);
    led_on();
    delay(100);
    led_off();
    delay(100);
    led_on();
    delay(100);
    led_off();
}

void wakeUP_RN2483() {
    Serial1.end();
    pinMode(PIN_SERIAL1_TX, OUTPUT);
    digitalWrite(PIN_SERIAL1_TX, LOW);
    delay(5);
    digitalWrite(PIN_SERIAL1_TX, HIGH);
    Serial1.begin(57600);
    Serial1.write(0x55);
    delay(100); //wait for the RN2xx3's startup message
    Serial1.flush();
}

// the loop routine runs over and over again forever:
void loop()
{
  if (count == 0) {
     //wakeUP_RN2483();
    myLora.autobaud();
    int16_t temp_c;
    int16_t humidity;
    int16_t battery;
    led_on();

    temp_c = int16_t(sht1x.readTemperatureC()*100);
    humidity = int16_t(sht1x.readHumidity()*100);
    battery = analogRead(batVoltPin)*4.88;


    txBuffer[0] = byte(temp_c>>8);
    txBuffer[1] = byte(temp_c);
    txBuffer[2] = byte(humidity>>8);
    txBuffer[3] = byte(humidity);
    txBuffer[4] = byte(battery>>8);
    txBuffer[5] = byte(battery);
    txBuffer[6] = 0xFE;

    myLora.txBytes(txBuffer, sizeof(txBuffer));
    led_off();
    myLora.sleep(128000);
    delay(100);
    count = 16;
  } else {
    LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
    count --;
  }
}

void led_on()
{
  digitalWrite(13, 1);
}

void led_off()
{
  digitalWrite(13, 0);
}

