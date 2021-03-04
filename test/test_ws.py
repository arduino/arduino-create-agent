import asyncio
import socketio as io
import time

def test_ws_connection(agent, socketio):
    print('my sid is', socketio.sid)
    assert socketio.sid is not None

def test_list(agent, socketio):
    socketio.emit('command', 'list')
    
