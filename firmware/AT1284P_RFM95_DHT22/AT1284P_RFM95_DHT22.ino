/*******************************************************************************
 * Copyright (c) 2017 JP Meijers, 
 * based on work by NC Thompson,
 * based on examples from the LMIC library.
 *
 * Permission is hereby granted, free of charge, to anyone
 * obtaining a copy of this document and accompanying files,
 * to do whatever they want with them without any restriction,
 * including, but not limited to, copying, modification and redistribution.
 * NO WARRANTY OF ANY KIND IS PROVIDED.
 * 
 * JP Meijers
 * 2017-10-03
 * 
 * Depends on:
 * https://github.com/JChristensen/mighty-1284p/tree/v1.6.3
 * https://github.com/matthijskooijman/arduino-lmic
 * https://github.com/lowpowerlab/lowpower
 * http://playground.arduino.cc/Main/DHTLib
 * 
 *******************************************************************************/

#include <lmic.h>
#include <hal/hal.h>
#include <SPI.h>
#include <LowPower.h>
#include "dht.h"

dht DHT;

#define DHT22_PIN 11
#define DHT_VCC 27

// LoRaWAN NwkSKey, network session key
static const u1_t NWKSKEY[16] = { 0xD4, 0x99, 0xE5, 0x84, 0x58, 0x24, 0x69, 0x17, 0xFD, 0x01, 0xFE, 0x8F, 0xBC, 0xA7, 0xDC, 0xA0 };

// LoRaWAN AppSKey, application session key
static const u1_t APPSKEY[16] = { 0x14, 0xEA, 0x44, 0x24, 0x13, 0x1E, 0xFF, 0x92, 0x30, 0x1C, 0x15, 0x6B, 0x75, 0x71, 0x5F, 0x05 };

// LoRaWAN end-device address (DevAddr)
static const u4_t DEVADDR = 0x2601186F;

// These callbacks are only used in over-the-air activation, so they are
// left empty here (we cannot leave them out completely unless
// DISABLE_JOIN is set in config.h, otherwise the linker will complain).
void os_getArtEui (u1_t* buf) { }
void os_getDevEui (u1_t* buf) { }
void os_getDevKey (u1_t* buf) { }

static uint8_t payloadSize = 7;
static uint8_t txBuffer[] = "0123456789";
static osjob_t sendjob;

// Time between samples in milliseconds
const unsigned TX_INTERVAL = 60000;
unsigned long startTime = 0;
unsigned long toSleep = 0;

int16_t temp_c = 0;
uint16_t humidity = 0;
int16_t battery = 0;

// Pin mapping
const lmic_pinmap lmic_pins = {
    .nss = 4,
    .rxtx = LMIC_UNUSED_PIN,
    .rst = 3,
    .dio = {2, 1, 0},
};

void onEvent (ev_t ev) {
    Serial.print(os_getTime());
    Serial.print(": ");
    switch(ev) {
        case EV_SCAN_TIMEOUT:
            Serial.println(F("EV_SCAN_TIMEOUT"));
            break;
        case EV_BEACON_FOUND:
            Serial.println(F("EV_BEACON_FOUND"));
            break;
        case EV_BEACON_MISSED:
            Serial.println(F("EV_BEACON_MISSED"));
            break;
        case EV_BEACON_TRACKED:
            Serial.println(F("EV_BEACON_TRACKED"));
            break;
        case EV_JOINING:
            Serial.println(F("EV_JOINING"));
            break;
        case EV_JOINED:
            Serial.println(F("EV_JOINED"));
            break;
        case EV_RFU1:
            Serial.println(F("EV_RFU1"));
            break;
        case EV_JOIN_FAILED:
            Serial.println(F("EV_JOIN_FAILED"));
            break;
        case EV_REJOIN_FAILED:
            Serial.println(F("EV_REJOIN_FAILED"));
            break;
            break;
        case EV_TXCOMPLETE:
            Serial.println(F("EV_TXCOMPLETE (includes waiting for RX windows)"));
            if(LMIC.dataLen) {
                // data received in rx slot after tx
                Serial.print(F("Data Received: "));
                Serial.write(LMIC.frame+LMIC.dataBeg, LMIC.dataLen);
                Serial.println();
            }
            break;
        case EV_LOST_TSYNC:
            Serial.println(F("EV_LOST_TSYNC"));
            break;
        case EV_RESET:
            Serial.println(F("EV_RESET"));
            break;
        case EV_RXCOMPLETE:
            // data received in ping slot
            Serial.println(F("EV_RXCOMPLETE"));
            break;
        case EV_LINK_DEAD:
            Serial.println(F("EV_LINK_DEAD"));
            break;
        case EV_LINK_ALIVE:
            Serial.println(F("EV_LINK_ALIVE"));
            break;
         default:
            Serial.println(F("Unknown event"));
            break;
    }
}

void do_send(osjob_t* j){
    // Check if there is not a current TX/RX job running
    if (LMIC.opmode & OP_TXRXPEND) {
        Serial.println(F("OP_TXRXPEND, not sending"));
    } else {
        // Prepare upstream data transmission at the next possible time.
        LMIC_setTxData2(1, txBuffer, payloadSize, 0);
        Serial.println(F("Packet queued"));
    }
    // Next TX is scheduled after TX_COMPLETE event.
}

void setup() {
  Serial.begin(115200);
  Serial.println(F("Starting"));
  
  pinMode(DHT_VCC, OUTPUT);
  pinMode(DHT22_PIN, INPUT);
  
  // LMIC init
  os_init();
  lmic_init();
}

void lmic_init()
{
    // LMIC init
    LMIC_reset();
    LMIC_setSession (0x1, DEVADDR, NWKSKEY, APPSKEY);

    LMIC_setupChannel(0, 868100000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(1, 868300000, DR_RANGE_MAP(DR_SF12, DR_SF7B), BAND_CENTI);      // g-band
    LMIC_setupChannel(2, 868500000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(3, 867100000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(4, 867300000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(5, 867500000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(6, 867700000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(7, 867900000, DR_RANGE_MAP(DR_SF12, DR_SF7),  BAND_CENTI);      // g-band
    LMIC_setupChannel(8, 868800000, DR_RANGE_MAP(DR_FSK,  DR_FSK),  BAND_MILLI);      // g2-band

    LMIC_setupBand (BAND_MILLI, 14, 1);
    LMIC_setupBand (BAND_CENTI, 14, 1);

    // Disable link check validation
    LMIC_setLinkCheckMode(0);

    // Set data rate and transmit power
    LMIC_setDrTxpow(DR_SF7,14);
}

void buildPayload() {
  battery = readVcc();
  
  // Switch on DHT22
  digitalWrite(DHT_VCC, HIGH);
  delay(1000);
  DHT.read22(DHT22_PIN);

  temp_c = int16_t(DHT.temperature*100);
  humidity = int16_t(DHT.humidity*100);
  digitalWrite(DHT_VCC, LOW);

  Serial.println(temp_c);
  Serial.println(humidity);
  Serial.println(battery);

  txBuffer[0] = byte(temp_c>>8);
  txBuffer[1] = byte(temp_c);
  txBuffer[2] = byte(humidity>>8);
  txBuffer[3] = byte(humidity);
  txBuffer[4] = byte(battery>>8);
  txBuffer[5] = byte(battery);
  txBuffer[6] = 0xFE;
  
}

void loop() {
  startTime = millis();
  buildPayload();
  // Queue packet for TX
  do_send(&sendjob);
  while(LMIC.opmode & OP_TXRXPEND)
  {
    os_runloop_once();
  }

  Serial.println("S");
  Serial.flush();
  toSleep = TX_INTERVAL - (millis() - startTime);
  if(toSleep > TX_INTERVAL) toSleep=60000; //Handle millis overflow
  sleep();
  Serial.println("W");
}

void sleep() {
  while(toSleep>8000) {
    LowPower.powerDown(SLEEP_8S, ADC_OFF, BOD_OFF);
    toSleep = toSleep - 8000;
  }
}

uint16_t readVcc() {
  // Read 1.1V reference against AVcc
  ADMUX = _BV(REFS0) | _BV(MUX4) | _BV(MUX3) | _BV(MUX2) | _BV(MUX1);

  delay(2); // Wait for Vref to settle
  ADCSRA |= _BV(ADSC); // Start conversion
  while (bit_is_set(ADCSRA,ADSC)); // measuring

  uint8_t low  = ADCL; // must read ADCL first - it then locks ADCH  
  uint8_t high = ADCH; // unlocks both

  uint16_t result = (high<<8) | low;

  result = 1125300L / result; // Calculate Vcc (in mV); 1125300 = 1.1*1023*1000
  return result; // Vcc in millivolts
}
