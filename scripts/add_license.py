import os

from pathlib import Path

# Apache-2.0 License header
license_header = '''\
// SPDX-License-Identifier: Apache-2.0
'''

def add_license_header(file_path) -> bool:
    with open(file_path, 'r+') as f:
        content = f.read()

        if license_header in content:
            return False

        f.seek(0, 0)
        f.write(license_header.lstrip('\n') + '\n' + content)
        return True


for file in Path(".").rglob("*.go"):
    add_license_header(file)
    print(f"-> Added license header to {file}")
