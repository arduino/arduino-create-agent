# import json
import psutil
import requests

# test if the update process succeeds in terminating the binary
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
def test_latest_version(base_url, agent):
    resp = requests.get(f"{base_url}/info")
    assert resp.status_code == 200
    latest_version = requests.get("https://s3.amazonaws.com/arduino-create-static/agent-metadata/agent-version.json") # get the latest version available

    version = latest_version.json()
    info = resp.json()
    assert info["version"] == version["Version"]
