from machine import PWM, Pin
import mqtt_reg

import sys
sys.path.append('/')
import site_config

print('Fan switch starting')

dutyRegister = None
enabledRegister = None

fan_pwm = PWM(Pin(site_config.fan_pin, Pin.OUT))
fan_pwm.init(freq=50000, duty_u16=0)

def update():
    duty = dutyRegister != None and enabledRegister != None and enabledRegister.get_value() and dutyRegister.get_value() or 0
    print('Fan duty:', duty)
    fan_pwm.duty_u16(int(duty * 65535))

update()

class DutyRegister(mqtt_reg.ClientRegister):
    def __init__(self):
        super().__init__(site_config.name + '.fan.duty')

    def set_value(self, value):
        if value < 0:
            value = 0
        if value > 1:
            value = 1
        super().set_value(value)
        update()


class EnabledRegister(mqtt_reg.BooleanPersistentServerRegister):
    def __init__(self):
        super().__init__(
            site_config.name + '.fan.enabled',
            {
                'device': site_config.name,
                'title': 'Fan switch master enable',
                'type': 'boolean'
            }, default=True)

    def set_value(self, value):
        super().set_value(value)
        update()


dutyRegister = DutyRegister()
enabledRegister = EnabledRegister()

registry = mqtt_reg.Registry(
    wifi_ssid=site_config.wifi_ssid,
    wifi_password=site_config.wifi_password,
    mqtt_broker=site_config.mqtt_broker,
    server=[enabledRegister],
    client=[dutyRegister],
    ledPin=site_config.wifi_led_pin,
    debug=site_config.debug
)

registry.start()
