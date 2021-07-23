import time
import json
import base64
import pytest

from common import running_on_ci
message = []

def test_ws_connection(socketio):
    print('my sid is', socketio.sid)
    assert socketio.sid is not None

def test_list(socketio):
    global message
    socketio.on('message', message_handler)
    socketio.emit('command', 'list')
    time.sleep(.1)
    # print (message)
    assert "list" in message[0]
    assert "Ports" in message[1]
    assert "Network" in message[2]

# NOTE run the following tests on linux with a board connected to the PC and with this sketch on it: https://gist.github.com/Protoneer/96db95bfb87c3befe46e
@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test__open_serial_default(socketio):
    global message
    message = []
    socketio.on('message', message_handler)
    socketio.emit('command', 'open /dev/ttyACM0 9600')
    time.sleep(.1) # give time to message to be filled
    assert "\"IsOpen\": true" in message[2]
    socketio.emit('command', 'send /dev/ttyACM0 /"ciao/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"ciao/\"" in message[3]
    assert "ciao" in extract_serial_data(message)

    # test with a lot of emoji: they can be messed up
    message = [] # reinitialize the message buffer
    socketio.emit('command', 'send /dev/ttyACM0 /"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in message[0]
    emoji_output = extract_serial_data(message)
    assert "/\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in emoji_output # this is failing because of UTF8 encoding problems 
    message = []
    socketio.emit('command', 'close /dev/ttyACM0')
    time.sleep(.1)
    # print (message)
    assert "\"IsOpen\": false," in message[8]

@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test__open_serial_timed(socketio):
    global message
    message = []
    socketio.on('message', message_handler)
    socketio.emit('command', 'open /dev/ttyACM0 9600 timed')
    time.sleep(.1) # give time to message to be filled
    assert "\"IsOpen\": true" in message[2]
    socketio.emit('command', 'send /dev/ttyACM0 /"ciao/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"ciao/\"" in message[3]
    assert "ciao" in extract_serial_data(message)

    # test with a lot of emoji: usually they get messed up
    message = [] # reinitialize the message buffer
    socketio.emit('command', 'send /dev/ttyACM0 /"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in message[0]
    emoji_output = extract_serial_data(message)
    assert "/\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in emoji_output
    message = []
    socketio.emit('command', 'close /dev/ttyACM0')
    time.sleep(.1)
    # print (message)
    assert "\"IsOpen\": false," in message[8]

@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test__open_serial_timedraw(socketio):
    global message
    message = []
    socketio.on('message', message_handler)
    socketio.emit('command', 'open /dev/ttyACM0 9600 timedraw')
    time.sleep(.1) # give time to message to be filled
    assert "\"IsOpen\": true" in message[2]
    socketio.emit('command', 'send /dev/ttyACM0 /"ciao/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"ciao/\"" in message[3]
    assert "ciao" in decode_output(extract_serial_data(message))

    # test with a lot of emoji: usually they get messed up
    message = [] # reinitialize the message buffer
    socketio.emit('command', 'send /dev/ttyACM0 /"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in message[0]
    # print (message)
    assert "/\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in decode_output(extract_serial_data(message))
    socketio.emit('command', 'close /dev/ttyACM0')
    time.sleep(.1)
    # print (message)
    assert "\"IsOpen\": false," in message[10]

@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test__open_serial_timedbinary(socketio):
    global message
    message = []
    socketio.on('message', message_handler)
    socketio.emit('command', 'open /dev/ttyACM0 9600 timedbinary')
    time.sleep(1) # give time to message to be filled
    assert "\"IsOpen\": true" in message[2]
    socketio.emit('command', 'send /dev/ttyACM0 /"ciao/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"ciao/\"" in message[3]
    assert "ciao" in decode_output(extract_serial_data(message))

    # test with a lot of emoji: usually they get messed up
    message = [] # reinitialize the message buffer
    socketio.emit('command', 'send /dev/ttyACM0 /"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/"')
    time.sleep(.1)
    assert "send /dev/ttyACM0 /\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in message[0]
    assert "/\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in decode_output(extract_serial_data(message))
    socketio.emit('command', 'close /dev/ttyACM0')
    time.sleep(.1)
    # print (message)
    assert "\"IsOpen\": false," in message[10]


# callback  called by socketio when a message is received
def message_handler(msg):
    # print('Received message: ', msg)
    global message
    message.append(msg)

# helper function used to extract serial data from it's json representation
# NOTE make sure to pass a clean message (maybe reinitialize the message global var before populating it)
def extract_serial_data(msg):
    serial_data = ""
    for i in msg:
        if "{\"P\"" in i:
            # print (json.loads(i)["D"])
            serial_data+=json.loads(i)["D"]
    # print("serialdata:"+serial_data)
    return serial_data
    
def decode_output(raw_output):
    # print(raw_output)
    base64_bytes = raw_output.encode('ascii') #encode rawoutput message into a bytes-like object
    output_bytes = base64.b64decode(base64_bytes)
    return output_bytes.decode('utf-8')
