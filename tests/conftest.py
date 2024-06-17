# Copyright 2022 Arduino SA
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.

import os
import platform
import signal
import time
from pathlib import Path

import pytest
from invoke import Local
from invoke.context import Context
import socketio as io

@pytest.fixture(scope="function")
def agent(pytestconfig):
    if platform.system() == "Windows":
        agent = str(Path(pytestconfig.rootdir) / "arduino-cloud-agent_cli.exe")
    else:
        agent = str(Path(pytestconfig.rootdir) / "arduino-cloud-agent")
    env = {
        # "ARDUINO_DATA_DIR": data_dir,
        # "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        # "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    run_context = Context()

    runner = Local(run_context) # execute a command on the local filesystem
    
    cd_command = "cd"
    with run_context.prefix(f'{cd_command} ..'):
        runner.run(agent, echo=True, hide=True, warn=True, env=env, asynchronous=True)
        
        # we give some time to the agent to start and listen to
        # incoming requests
        time.sleep(1)

        # we block here until the test function using this fixture has returned
        yield runner

    # Kill the runner's process as we finished our test (platform dependent)
    os_signal = signal.SIGTERM
    if platform.system() != "Windows":
        os_signal = signal.SIGKILL
    os.kill(runner.process.pid, os_signal)


@pytest.fixture(scope="session")
def base_url():
    return "http://127.0.0.1:8991"

@pytest.fixture(scope="function")
def socketio(base_url, agent):
    sio = io.Client()
    sio.connect(base_url)
    yield sio
    sio.disconnect()

@pytest.fixture(scope="session")
def serial_port():
    return "/dev/ttyACM0" # maybe this could be enhanced by calling arduino-cli

@pytest.fixture(scope="session")
def baudrate():
    return "9600"

# open_port cannot be coced as a fixture because of the buffertype parameter

# at the end of the test closes the serial port
@pytest.fixture(scope="function")
def close_port(socketio, serial_port):
    yield socketio
    socketio.emit('command', 'close ' + serial_port)
    time.sleep(.5)


@pytest.fixture(scope="function")
def message(socketio):
    global message
    message = []
    #in message var we will find the "response"
    socketio.on('message', message_handler)
    return message

# callback  called by socketio when a message is received
def message_handler(msg):
    # print('Received message: ', msg)
    global message
    message.append(msg)