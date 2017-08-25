package agent

import exec "github.com/arduino/arduino-create-agent/exec"

var boards = map[string]exec.Command{"upload:arduino:avr:one": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:lilypad:cpu=atmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:lilypad:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:samd:arduino_zero_native": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac-1.7.0.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -i -e -w -v \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:arduino:avr:atmegang:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:atmegang:cpu=atmega8": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega8 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:robotMotor": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:mini:cpu=atmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:mini:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:chiwawa": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:samd:arduino_zero_edbg": exec.Command{
	Params:  []string{"{upload.verbose}", "{build.path}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.openocd.path}/bin/openocd\" {upload.verbose} -s \"{runtime.tools.openocd.path}/share/openocd/scripts/\" -f \"{build.path}/arduino_zero.cfg\" -c \"telnet_port disabled; program {build.path}/{build.project_name}.bin verify reset 0x00002000; shutdown\"",
}, "upload:littleBits:avr:w6_arduino": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:megaADK": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega2560 -cwiring -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:circuitplay32u4cat": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:mega:cpu=atmega1280": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega1280 -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:mega:cpu=atmega2560": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega2560 -cwiring -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:leonardo": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:nrf52:primo": exec.Command{
	Params:  []string{"{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.openocd-0.10.0-arduino1-static.path}/bin/openocd\" -s \"{runtime.tools.openocd-0.10.0-arduino1-static.path}/share/openocd/scripts/\" -f \"{runtime.platform.path}/arduino_primo.cfg\" -c \"program {{{build.path}/{build.project_name}-merged.hex}} verify reset exit\"",
}, "upload:arduino:sam:arduino_due_x": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{upload.verify}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -e -w {upload.verify} -b \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:arduino:samd:mkr1000": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac-1.7.0.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -i -e -w -v \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:arduino:avr:robotControl": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:pro:cpu=8MHzatmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:pro:cpu=16MHzatmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:pro:cpu=8MHzatmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:pro:cpu=16MHzatmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:nano:cpu=atmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:nano:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:samd:mkrfox1200": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac-1.7.0.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -i -e -w -v \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:arduino:stm32f4:star_otto": exec.Command{
	Params:  []string{"{build.path}", "{build.project_name}", "{serial.port}", "{upload.params.verbose}"},
	Pattern: "\"{runtime.tools.arduinoSTM32load-2.0.0.path}/arduinoSTM32load\" -dfu \"{runtime.tools.dfu-util-0.9.0-arduino1.path}\" -bin \"{build.path}/{build.project_name}.bin\" -port=\"{serial.port}\" \"{upload.params.verbose}\"",
}, "upload:arduino:samd:mkrzero": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac-1.7.0.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -i -e -w -v \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:atmel-avr-xminis:avr:atmega328p_xplained_mini": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:gemma": exec.Command{
	Params:  []string{"{upload.verbose}", "{upload.protocol}", "{serial.port}", "{upload.speed}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -pattiny85 -c{upload.protocol} -P{serial.port} -b{upload.speed} -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:micro": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:unowifi": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:bt:cpu=atmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:bt:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:atmel-avr-xminis:avr:atmega328pb_xplained_mini": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:leonardoeth": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:diecimila:cpu=atmega328": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:diecimila:cpu=atmega168": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168 -carduino -P{serial.port} -b19200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:yun": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:ethernet": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:Intel:arc32:arduino_101": exec.Command{
	Params:  []string{"{build.path}", "{build.project_name}", "{serial.port}", "{upload.verbose}"},
	Pattern: "\"{runtime.tools.arduino101load.path}/arduino101load\" \"-dfu={runtime.tools.dfu-util.path}\" \"-bin={build.path}/{build.project_name}.bin\" -port={serial.port} \"{upload.verbose}\" -ble_fw_str=\"ATP1BLE00R-1631C4439\" -ble_fw_pos=169984 -rtos_fw_str=\"\" -rtos_fw_pos=0 -core=2.0.0",
}, "upload:arduino:avr:uno": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b115200 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:fio": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega328p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:sam:arduino_due_x_dbg": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{upload.verify}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac.path}/bossac\" {upload.verbose} --port={serial.port.file} -U false -e -w {upload.verify} -b \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:arduino:samd:adafruit_circuitplayground_m0": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port.file}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.bossac-1.7.0.path}/bossac\" {upload.verbose} --port={serial.port.file} -U true -i -e -w -v \"{build.path}/{build.project_name}.bin\" -R",
}, "upload:atmel-avr-xminis:avr:atmega168pb_xplained_mini": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega168p -carduino -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:yunmini": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:LilyPadUSB": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}, "upload:arduino:avr:esplora": exec.Command{
	Params:  []string{"{upload.verbose}", "{serial.port}", "{build.path}", "{build.project_name}"},
	Pattern: "\"{runtime.tools.avrdude.path}/bin/avrdude\" \"-C{runtime.tools.avrdude.path}/etc/avrdude.conf\" {upload.verbose}  -patmega32u4 -cavr109 -P{serial.port} -b57600 -D \"-Uflash:w:{build.path}/{build.project_name}.hex:i\"",
}}
