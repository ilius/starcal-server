#!/usr/bin/python3

import os
import sys

try:
	import Crypto
except:
	print("Crypto module was not found, try: sudo pip3 install PyCrypto")
	os.exit(1)

from Crypto.Cipher import AES
from Crypto import Random
import binascii

from base64 import b64encode

masterKeyHex = os.getenv("MASTER_KEY")
if not masterKeyHex:
	print("MASTER_KEY is not set")
	os.exit(1)

masterKey = binascii.unhexlify(masterKeyHex)

print("Type or paste the secret value and press Enter:")

secret = sys.stdin.readline()
secret = secret.rstrip("\n")
print(repr(secret))

iv = Random.new().read(AES.block_size)
cipher = AES.new(masterKey, AES.MODE_CBC, iv)
secretEncrypted = iv + cipher.encrypt(secret)

print()
print(b64encode(secretEncrypted).decode())