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
    time.sleep(.2)
    print (message)
    assert any("list" in i for i in message)
    assert any("Ports" in i for i in message)
    assert any("Network" in i for i in message)

# NOTE run the following tests with a board connected to the PC
@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_open_serial_default(socketio, serial_port, baudrate):
    general_open_serial(socketio, serial_port, baudrate, "default")


@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_open_serial_timed(socketio, serial_port, baudrate):
    general_open_serial(socketio, serial_port, baudrate, "timed")


@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_open_serial_timedraw(socketio, serial_port, baudrate):
    general_open_serial(socketio, serial_port, baudrate, "timedraw")


# NOTE run the following tests with a board connected to the PC and with the sketch found in test/testdata/SerialEcho.ino on it be sure to change serial_address in conftest.py
@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_send_serial_default(socketio, close_port, serial_port, baudrate):
    general_send_serial(socketio, close_port, serial_port, baudrate, "default")


@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_send_serial_timed(socketio, close_port, serial_port, baudrate):
    general_send_serial(socketio, close_port, serial_port, baudrate, "timed")


@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_send_serial_timedraw(socketio, close_port, serial_port, baudrate):
    general_send_serial(socketio, close_port, serial_port, baudrate, "timedraw")


def general_open_serial(socketio, serial_port, baudrate, buffertype):
    global message
    message = []
    # in message var we will find the "response"
    socketio.on('message', message_handler)
    socketio.emit('command', 'open ' + serial_port + ' ' + baudrate + ' ' + buffertype)
    # give time to the message var to be filled
    time.sleep(.5)
    print(message)
    # the serial connection should be open now
    assert any("\"IsOpen\": true" in i for i in message)

    # close the serial port
    socketio.emit('command', 'close ' + serial_port)
    time.sleep(.2)
    print (message)
    #check if port has been closed
    assert any("\"IsOpen\": false," in i for i in message)



def general_send_serial(socketio, close_port, serial_port, baudrate, buffertype):
    global message
    message = []
    #in message var we will find the "response"
    socketio.on('message', message_handler)
    #open a new serial connection with the specified buffertype, if buffertype is empty it will use the default one
    socketio.emit('command', 'open ' + serial_port + ' ' + baudrate + ' ' + buffertype)
    # give time to the message var to be filled
    time.sleep(.5)
    print(message)
    # the serial connection should be open now
    assert any("\"IsOpen\": true" in i for i in message)

    #test with string
    # send the string "ciao" using the serial connection
    socketio.emit('command', 'send ' + serial_port + ' /"ciao/"')
    time.sleep(1)
    print(message)
    # check if the send command has been registered
    assert any("send " + serial_port + " /\"ciao/\"" in i for i in message)
    #check if message has been sent back by the connected board
    if buffertype == "timedraw":
        output =  decode_output(extract_serial_data(message))
    elif buffertype in ("default", "timed"):
        output = extract_serial_data(message)
    assert "ciao" in output

    #test with emoji
    message = [] # reinitialize the message buffer to have a clean situation
    # send a lot of emoji: they can be messed up
    socketio.emit('command', 'send ' + serial_port + ' /"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/"')
    time.sleep(.5)
    print(message)
    # check if the send command has been registered
    assert any("send " + serial_port + " /\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in i for i in message)
    if buffertype == "timedraw":
        output =  decode_output(extract_serial_data(message))
    elif buffertype in ("default", "timed"):
        output = extract_serial_data(message)
    assert "/\"🧀🧀🧀🧀🧀🧀🧀🧀🧀🧀/\"" in output
    # the serial connection is closed by close_port() fixture: even if in case of test failure

@pytest.mark.skipif(
    running_on_ci(),
    reason="VMs have no serial ports",
)
def test_sendraw_serial(socketio, close_port, serial_port, baudrate):
    global message
    message = []
    #in message var we will find the "response"
    socketio.on('message', message_handler)
    #open a new serial connection with the specified buffertype, if buffertype is empty it will use the default one
    socketio.emit('command', 'open ' + serial_port + ' ' + baudrate + ' timedraw')
    # give time to the message var to be filled
    time.sleep(.5)
    print(message)
    # the serial connection should be open now
    assert any("\"IsOpen\": true" in i for i in message)

    #test with bytes
    integers = [1, 2, 3, 4, 5]
    bytes_array=bytearray(integers)
    encoded_integers = base64.b64encode(bytes_array).decode('ascii')
    socketio.emit('command', 'sendraw ' + serial_port + ' ' + encoded_integers)
    time.sleep(1)
    print(message)
    # check if the send command has been registered
    assert any(("sendraw " + serial_port + ' ' + encoded_integers) in i for i in message)
    #check if message has been sent back by the connected board
    output =  extract_serial_data(message) # TODO use decode_output()
    print (output)
    assert encoded_integers in output


# callback  called by socketio when a message is received
def message_handler(msg):
    # print('Received message: ', msg)
    global message
    message.append(msg)

# helper function used to extract serial data from its JSON representation
# NOTE make sure to pass a clean message (maybe reinitialize the message global var before populating it)
def extract_serial_data(msg):
    serial_data = ""
    for i in msg:
        if "{\"P\"" in i:
            print (json.loads(i)["D"])
            serial_data+=json.loads(i)["D"]
    print("serialdata:"+serial_data)
    return serial_data
    
def decode_output(raw_output):
    # print(raw_output)
    base64_bytes = raw_output.encode('ascii') #encode rawoutput message into a bytes-like object
    output_bytes = base64.b64decode(base64_bytes)
    return output_bytes.decode('utf-8')
