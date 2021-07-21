import time

def test_ws_connection(socketio):
    print('my sid is', socketio.sid)
    assert socketio.sid is not None

def test_list(socketio):
    socketio.on('message', message_handler)
    socketio.emit('command', 'list')
    time.sleep(.1)

def test__open_serial_default(socketio):
    socketio.on('message', message_handler)
    socketio.emit('command', 'open /dev/ttyACM0 9600')
    time.sleep(.1)
    socketio.emit('command', 'send /dev/ttyACM0 /"ciao/"')
    time.sleep(.1)
    socketio.emit('command', 'close /dev/ttyACM0')
    time.sleep(.1)


def message_handler(msg):
    print('Received message: ', msg)