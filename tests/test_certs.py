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
