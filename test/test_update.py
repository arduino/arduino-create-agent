import json
import requests


def test_update_shutdown(base_url, agent):
    
    resp = requests.post(f"{base_url}/update")
    assert resp.status_code == 200
    info = resp.json()
    assert "Please wait a moment while the agent reboots itself" in info['success']

