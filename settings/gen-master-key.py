#!/usr/bin/python3

import os
import binascii

# length can be 16, 24 or 32
length = 32

print(binascii.hexlify(os.urandom(length)).decode())

## In Python 3.6 you can use secrets.token_hex(length)