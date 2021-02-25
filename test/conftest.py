import os
import platform
import signal
import time
from pathlib import Path

import pytest
from invoke import Local
from invoke.context import Context


@pytest.fixture(scope="function")
def agent(pytestconfig):
    
    agent_cli = str(Path(pytestconfig.rootdir) / "arduino-create-agent_cli")
    env = {
        # "ARDUINO_DATA_DIR": data_dir,
        # "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        # "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    run_context = Context()

    runner = Local(run_context) # execute a command on the local filesystem
    
    cd_command = "cd"
    if platform.system() == "Windows":
        cd_command += " /d"

    with run_context.prefix(f'{cd_command} ..'):
        runner.run(agent_cli, echo=True, hide=True, warn=True, env=env, asynchronous=True)
        
        # we give some time to the agent to start and listen to
        # incoming requests
        time.sleep(.5)

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
