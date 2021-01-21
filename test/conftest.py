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
    
    cli_full_line = str(Path(pytestconfig.rootdir) / "arduino-create-agent")
    env = {
        # "ARDUINO_DATA_DIR": data_dir,
        # "ARDUINO_DOWNLOADS_DIR": downloads_dir,
        # "ARDUINO_SKETCHBOOK_DIR": data_dir,
    }
    run_context = Context()

    # TODO: wtf is this?
    runner = Local(run_context)
    
    cd_command = "cd"
    with run_context.prefix(f'{cd_command} ..'):
        runner.run(cli_full_line, echo=True, hide=True, warn=True, env=env, asynchronous=True)
        
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
