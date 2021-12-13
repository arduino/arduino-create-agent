import re
import requests


def test_version(base_url, agent):
    
    resp = requests.get(f"{base_url}/info")
    assert resp.status_code == 200

    info = resp.json()
    assert re.match("[0-9]+.[0-9]+.[0-9]+", info["version"]) is not None
