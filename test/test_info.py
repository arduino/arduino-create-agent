import re
import requests


def test_info(base_url, agent):
    
    resp = requests.get(f"{base_url}/info")
    assert resp.status_code == 200
    info = resp.json()
    assert re.match("[0-9]+.[0-9]+.[0-9]+", info["version"]) is not None
    assert info["http"] is not None
    assert info["https"] is not None
    assert info["origins"] is not None
    assert info["os"] is not None
    assert info["update_url"] is not None
    assert info["ws"] is not None
    assert info["wss"] is not None
