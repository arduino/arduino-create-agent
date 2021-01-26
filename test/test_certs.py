import requests


def test_get_cert(base_url, agent):
    
    resp = requests.get(f"{base_url}/certificate.crt")
    assert resp.status_code == 200

    cert = resp.text
    assert "<!DOCTYPE html>" in cert


def test_del_cert(base_url, agent):
    
    resp = requests.delete(f"{base_url}/certificate.crt")
    assert resp.status_code == 200

    # Should rm "ca.cert.pem", "ca.cert.cer", "ca.key.pem"
