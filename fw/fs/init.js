/* global RegisterGPO, Register, RegisterDS18B20 */

load('api_config.js');
load("api_gpio.js");
load("api_timer.js");

load("api_df_reg.js");
load("api_df_reg_gpo.js");
load("api_df_reboot.js");

Reboot.after(10);

let gateName = Cfg.get("fan.gate.name");
if (gateName) {
    let gatePin = Cfg.get("fan.gate.pin");
    print("Gate register", gateName, "at pin", gatePin);
    let defValue = Cfg.get("fan.gate.def");
    let register = RegisterGPO.create(gatePin);
    register.set(defValue);
    Register.add(gateName, register);
}

