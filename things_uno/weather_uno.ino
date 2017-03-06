/*
 * Authors: NC Thompson, JP Meijers
 * Date: 2017-02-15
 *
 */
#include <rn2xx3.h>
#include <SHT1x.h>
#include <avr/wdt.h>
#include <avr/power.h>
#include <avr/sleep.h>

// The interval at which the sensors should be read and the values transmitted
#define SAMPLE_INTERVAL_MS 60000UL

// after this amount of time, if the rn2483 did not wake us up, reset the system
#define WDT_INTERVAL_MS 600000UL

//create an instance of the rn2483 library, using the given Serial port
rn2xx3 myLora(Serial1);
#define dataPin  SDA
#define clockPin SCL
#define SELF_RESET_PIN 12

SHT1x sht1x(dataPin, clockPin);

uint8_t txBuffer[7];
int batVoltPin = A5;    // select the input pin for the battery voltage

unsigned long startTime = 0;
unsigned long sleepTime = 0;
unsigned int wdtCount = 0;


/*
 * Define Timer4 labels
 */
#if defined __AVR_ATmega32U4__
  // Timer 4 PRR bit is currently not defined in iom32u4.h
  #ifndef PRTIM4
    #define PRTIM4 4
  #endif

  // Timer 4 power reduction macro is not defined currently in power.h
  #ifndef power_timer4_disable
    #define power_timer4_disable()  (PRR1 |= (uint8_t)(1 << PRTIM4))
  #endif

  #ifndef power_timer4_enable
    #define power_timer4_enable()   (PRR1 &= (uint8_t)~(1 << PRTIM4))
  #endif
#endif



void setup()
{
  digitalWrite(SELF_RESET_PIN, HIGH);
  wdt_disable();

  //output LED pin
  pinMode(13, OUTPUT);
  led_on();

  Serial.begin(57600); //serial port to computer

  // make sure usb serial connection is available,
  // or after 10s go on anyway for 'headless' use of the node.
  while ((!Serial) && (millis() < 10000));

  Serial.println("Startup");

  initialize_radio();

  // set up interrupt on Serial1 RX
  // RXD1 = INT2 - enable interrupt on rising and falling edge
  EICRA = 0b01010101;
  EIMSK = 0b00000100;

  Serial.end();
  led_off();
}

void initialize_radio()
{
  Serial1.begin(57600); //serial port to radio

  //reset rn2483
  pinMode(4, OUTPUT);
  digitalWrite(4, LOW);
  delay(100);
  digitalWrite(4, HIGH);

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

// the loop routine runs over and over again forever:
void loop()
{
  startTime = millis();
  led_on();
  myLora.autobaud();
  led_off();
  delay(100);
  led_on();
  int16_t temp_c;
  int16_t humidity;
  int16_t battery;

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

  sleepTime = (SAMPLE_INTERVAL_MS-(millis()-startTime));
  myLora.sleep(sleepTime);
  delay(100);

  go_to_sleep();
}

void led_on()
{
  digitalWrite(13, 1);
}

void led_off()
{
  digitalWrite(13, 0);
}

ISR(WDT_vect)
{
  wdt_reset();
  wdtCount++;
  if(wdtCount<=(WDT_INTERVAL_MS/8000))
  {
    sleep_now();
  }
  else
  {
    pinMode(12, OUTPUT);
    digitalWrite(12, LOW);
  }
}

ISR(INT2_vect)
{
  //wake up if we are in sleep
}

ISR(BADISR_vect)
{
  //unknown interrupt - do nothing, just wake up
}

void enable_wdt_interrupt_8s()
{
  /*
  WDTCSR configuration:
  WDIE = 1 :Interrupt Enable
  WDE = 1  :Reset Enable
  WDP3:0 = 1001 = 8s
  */
  MCUSR &= ~(1<<WDRF);  // because the data sheet said to
  // Enter Watchdog Configuration mode:
  WDTCSR = (1<<WDCE) | (1<<WDE);
  // Set Watchdog settings: interrupte enable, 1001 for timer
  WDTCSR = (0<<WDE) | (1<<WDIE) | (1<<WDP3) | (0<<WDP2) | (0<<WDP1) | (1<<WDP0);
  wdt_reset();
}

void go_to_sleep()
{
  enable_wdt_interrupt_8s();

  ADCSRA &= ~(1 << ADEN);
  power_adc_disable();

  power_timer4_disable();
  power_timer3_disable();
  power_timer1_disable();
  power_timer0_disable();
  power_spi_disable();
  power_usart1_disable();
  power_twi_disable();
  power_usb_disable();

  sleep_now();

  wdt_reset();
  wdt_disable();
  wdtCount=0;

  power_adc_enable();
  ADCSRA |= (1 << ADEN);

  power_timer4_enable();
  power_timer3_enable();
  power_timer1_enable();
  power_timer0_enable();
  power_spi_enable();
  power_usart1_enable();
  power_twi_enable();
  power_usb_enable();
}

void sleep_now()
{
  // mcu sleep
  cli();
  set_sleep_mode(SLEEP_MODE_PWR_DOWN);
  sleep_enable();
  //sleep_bod_disable(); // Only Pico Power devices can change BOD settings through software
  sei();
  sleep_cpu();
  sleep_disable();
}

