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

import re
import requests
import pytest
from sys import platform


@pytest.mark.skipif(
    platform == "darwin",
    reason="on macOS the user is prompted to install certificates",
)
def test_version(base_url, agent):
    
    resp = requests.get(f"{base_url}/info")
    assert resp.status_code == 200

    info = resp.json()
    assert re.match("[0-9]+.[0-9]+.[0-9]+", info["version"]) is not None
