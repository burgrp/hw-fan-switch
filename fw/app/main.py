from machine import UART, Pin
import time
import mqtt_reg

import sys
sys.path.append('/')

import site_config

WIFI_LED_PIN = 12
FAN_PIN = 13

print('FAN switch starting')

fan = Pin(FAN_PIN, Pin.OUT)

class RegistryHandler:

    def __init__(self):
        pass

    def get_names(self):
        return [site_config.name + ".fan"]

    def get_meta(self, name):
        return {
            'device': site_config.name,
            'title': 'Fan switch state',
            'type': 'boolean'
        }

    def get_value(self, name):
        return fan.value() == 1

    def set_value(self, name, value):
        fan.value(1 if value else 0)

registry = mqtt_reg.Registry(
    RegistryHandler(),
    wifi_ssid=site_config.wifi_ssid,
    wifi_password=site_config.wifi_password,
    mqtt_broker=site_config.mqtt_broker,
    ledPin=WIFI_LED_PIN,
    debug=site_config.debug
)

registry.start()

