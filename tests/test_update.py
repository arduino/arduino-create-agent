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

# import json
import psutil
import requests
import pytest

# test if the update process succeeds in terminating the binary
@pytest.mark.skip(reason="no way of currently testing this")
def test_update_shutdown(base_url, agent):

    procs=[]
    for p in psutil.process_iter():
        if p.name() == "arduino-create-agent":
            procs.append(p)

    resp = requests.post(f"{base_url}/update")
    # assert resp.status_code == 200
    # assert "Please wait a moment while the agent reboots itself" in info['success'] # failing on macos see https://github.com/arduino/arduino-create-agent/issues/608
    gone, alive = psutil.wait_procs(procs, timeout=3, callback=on_terminate) # wait for "arduino-create-agent" to terminate

def on_terminate(proc):
    print("process {} terminated with exit code {}".format(proc, proc.returncode))
    assert True

# the version currently running is the latest available
@pytest.mark.skip(reason="no way of currently testing this")
def test_latest_version(base_url, agent):
    resp = requests.get(f"{base_url}/info")
    assert resp.status_code == 200
    latest_version = requests.get("https://s3.amazonaws.com/arduino-create-static/agent-metadata/agent-version.json") # get the latest version available

    version = latest_version.json()
    info = resp.json()
    assert info["version"] == version["Version"]
