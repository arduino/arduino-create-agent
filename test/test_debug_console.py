import requests
from bs4 import BeautifulSoup

def test_valid_html(base_url, agent):
    resp = requests.get(f"{base_url}")
    assert resp.status_code == 200
    print(resp.text)
    assert bool(BeautifulSoup(resp.text, "html.parser").find())

