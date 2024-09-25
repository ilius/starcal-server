#!/usr/bin/python3

import os
import sys

try:
	import Crypto  # noqa: F401
except ImportError:
	print("Crypto module was not found, try: sudo pip3 install PyCrypto")
	sys.exit(1)

import binascii
from base64 import b64encode

from Crypto import Random
from Crypto.Cipher import AES

masterKeyHex = os.getenv("MASTER_KEY")
if not masterKeyHex:
	print("MASTER_KEY is not set")
	sys.exit(1)

masterKey = binascii.unhexlify(masterKeyHex)

print("Type or paste the secret value and press Enter:")

secret = sys.stdin.readline()
secret = secret.rstrip("\n")
print(repr(secret))

toAppendBytes = 15 - (len(secret) - 1) % 16
if toAppendBytes > 0:
	secret += "\x00" * toAppendBytes

iv = Random.new().read(AES.block_size)
cipher = AES.new(masterKey, AES.MODE_CBC, iv)
secretEncrypted = iv + cipher.encrypt(secret)

print()
print(b64encode(secretEncrypted).decode())
