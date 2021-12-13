import requests


def test_get_tools(base_url, agent):
    
    resp = requests.get(f"{base_url}/v2/pkgs/tools/installed")
    assert resp.status_code == 200

    tools = resp.json()