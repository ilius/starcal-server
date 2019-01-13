#!/usr/bin/python3.6

import sys
import os
from os.path import dirname, join, abspath, isdir, isfile

import json
import requests

from datetime import datetime
from datetime import timedelta
from dateutil.parser import parse as parse_time

import dateutil.tz
from dateutil.tz.tz import tzutc

from parse import parse as parse_format


from lxml import etree
from lxml.etree import _Element as Element

# the code for cmd.Cmd is very ugly and hard to understan

# readline's complete func silently (and stupidly) hides any exception
# and only shows the print if it's in the first line of function. very awkward!

#import atexit

from prompt_toolkit import prompt
from prompt_toolkit.history import FileHistory
from prompt_toolkit.auto_suggest import AutoSuggestFromHistory
from prompt_toolkit.completion.word_completer import WordCompleter


from typing import Dict, Tuple, List, Optional, Any

utc = tzutc()


host = os.getenv("STARCAL_HOST")
if not host:
	import socket
	host = socket.gethostname()

secure = (os.getenv("STARCAL_HOST_SECURE", "") != "")

port = "9001"

baseURL = ("https" if secure else "http") + "://" + host + ":" + port

print("< Base URL:", baseURL)

myDir = dirname(__file__)
cwd = os.getcwd()
if myDir in (".", ""):
	myDir = cwd
elif os.sep == "/":
	if myDir.startswith("./"):
		myDir = cwd + myDir[1:]
	elif myDir[0] != "/":
		myDir = join(cwd, myDir)
elif os.sep == "\\":
	if myDir.startswith(".\\"):
		myDir = cwd + myDir[1:]
#print("myDir=%r"%myDir)

rootDir = abspath(dirname(myDir))
docPath = join(rootDir, "docs", "api-v1.wadl")

homeDir = os.getenv("HOME")
confDir = join(homeDir, ".starcal-server-cli")
histDir = join(confDir, "history")
tokenDirParent = join(confDir, "auth-tokens")
tokenDir = join(tokenDirParent, host)
lastPathFile = join(confDir, "last-path")

if not isdir(confDir):
	os.mkdir(confDir, 0o700)
if not isdir(histDir):
	os.mkdir(histDir, 0o700)
if not isdir(tokenDirParent):
	os.mkdir(tokenDirParent, 0o700)
if not isdir(tokenDir):
	os.mkdir(tokenDir, 0o700)

indent = "\t"

with open(docPath, encoding="utf-8") as docFile:
	doc = etree.XML(docFile.read().encode("utf-8"))

def dataToPrettyJson(data, ensure_ascii=False, sort_keys=False):
	return json.dumps(
		data,
		sort_keys=sort_keys,
		indent=2,
		ensure_ascii=ensure_ascii,
	)

def getEmail() -> str:
	email = os.getenv("STARCAL_EMAIL")
	if email:
		return email
	while True:
		email = prompt(
			"Email: ",
			history=FileHistory(join(histDir, "email")),
			auto_suggest=AutoSuggestFromHistory(),
		)
		if email:
			return email
	raise ValueError("email is not given")

def getPassword() -> str:
	password = os.getenv("STARCAL_PASSWORD")
	if password:
		return password
	import getpass
	while True:
		try:
			password = getpass.getpass("Password: ")
		except KeyboardInterrupt:
			return ""
		if password:
			return password
	return ""


tokenExpFormat = "%Y-%m-%dT%H:%M:%SZ"

localTZ = dateutil.tz.gettz()

# returns token
def getSavedToken(email: str) -> Optional[str]:
	tokenPath = join(tokenDir, email)
	if not isfile(tokenPath):
		print("Token file does not exist:", tokenPath)
		return
	with open(tokenPath) as tokenFile:
		tokenJson = tokenFile.read()
	tokenDict = json.loads(tokenJson)
	# print(tokenDict)
	expStr = tokenDict.get("expiration", "")
	if not expStr:
		print("WARNING: invalid token file %r, no 'expiration'" % tokenPath)
		return
	token = tokenDict.get("token", "")
	if not token:
		print("WARNING: invalid token file %r, no 'token'" % tokenPath)
		return
	exp = parse_time(expStr)
	if exp - timedelta(minutes=5) < datetime.now(tz=localTZ):
		print("Saved token in %s is expired or about to be expired on %s" % (tokenPath, exp))
		return
	return token

# returns (email, token), error
def getAuth() -> Tuple[Optional[Tuple[str, str]], Optional[str]]:
	email = getEmail()
	token = getSavedToken(email)
	if token:
		return (email, token), None
	password = getPassword()
	if not password:
		return None, "password is empty"
	url = baseURL + "/auth/login/"
	print("Sending login request")
	res = requests.post(url, json={
		"email": email,
		"password": password,
	})
	err = None
	try:
		resData = res.json()
	except:
		resData = None
	else:
		err = resData.get("error")
	if res.status_code != 200:
		if not err:
			err = "%s from %s" % (res, url)
		return None, err
	if resData is None:
		return "", "ERROR: non-json body from %r" % url
	token = resData.get("token")
	if not token:
		return None, "login returned no token"

	try:
		expStr = resData["expiration"]
	except KeyError:
		return None, "login returned no expiration"

	tokenPath = join(tokenDir, email)
	with open(tokenPath, "w") as tokenFile:
		json.dump({
			"token": token,
			"expiration": expStr,
		}, tokenFile)

	return (email, token), None


def getElemTag(elem) -> str:
	# for example, elem.tag == "{http://localhost:9001/application.wadl}resource"
	return elem.tag.split("}")[1]

def elemRepr(elem) -> str:
	tag = getElemTag(elem)
	prefix = indent * level + getElemTag(elem)
	if tag == "resource":
		return elem.get("path", "NO_PATH")
	elif tag == "method":
		return elem.get("name", "NO_NAME") + " (" + elem.get("id", "NO_ID") + ")"
	elif tag == "param" or tag == "element":
		return elem.get("name", "NO_NAME") + " (type=" + elem.get("type", "NO_ID") + ")"
	elif tag == "item":
		return "(type=" + elem.get("type", "NO_TYPE") + ")"
	elif tag == "option":
		return elem.get("value", "NO_VALUE")
	elif tag == "representation":
		return ""
	return tag

def printElem(elem: Element, level: int):
	tag = getElemTag(elem)
	prefix = indent * level + getElemTag(elem)
	if tag == "resource":
		print(prefix + ": " + elem.get("path", "NO_PATH"))
	elif tag == "method":
		print(prefix + ": " + elem.get("name", "NO_NAME") + " (" + elem.get("id", "NO_ID") + ")")
	elif tag == "param" or tag == "element":
		print(prefix + ": " + elem.get("name", "NO_NAME") + " (type=" + elem.get("type", "NO_ID") + ")")
	elif tag == "item":
		print(prefix + " (type=" + elem.get("type", "NO_TYPE") + ")")
	elif tag == "option":
		print(prefix + ": " + elem.get("value", "NO_VALUE"))
	elif tag == "representation":
		pass
	else:
		print(prefix)
	for child in elem.getchildren():
		printElem(child, level+1)

def nonEmptyStrings(*args) -> List[str]:
	l = [] # type: List[str]
	for x in args:
		if x:
			l.append(x)
	return l

def elemKeys(elem) -> str:
	# returns virtual file names of element
	tag = getElemTag(elem)
	if tag == "resource":
		return nonEmptyStrings(elem.get("path", None))
	elif tag == "method" or tag == "param" or tag == "element":
		return nonEmptyStrings(elem.get("name", None), elem.get("id", None))
	# elif tag == "item":
	# 	return [] # FIXME
	# elif tag == "option":
	# 	return nonEmptyStrings(elem.get("value", None))
	elif tag == "representation":
		return []
	else:
		return []
		# print("elemPath", prefix)

# returns (options, optionsMinimal)
def elemChildOptions(elem: Element) -> Tuple[Dict[str, Element], Dict[str, Element]]:
	options = {} # type: Dict[str, Element]
	optionsMinimal = {} # type: Dict[str, Element]
	for child in elem.getchildren():
		keys = elemKeys(child)
		if not keys:
			continue
		optionsMinimal[keys[0]] = child
		for key in keys:
			options[key] = child
			options[key.lower()] = child
	return options, optionsMinimal


def getParamCompleter(elem: Element) -> Optional[WordCompleter]:
	options = [] # type: List[str]
	for child in elem.getchildren():
		if getElemTag(child) == "option":
			value = child.get("value", None)
			if value:
				options.append(value)
	if not options:
		return None
	return WordCompleter(
		options,
		ignore_case=False,
	)

def getMethodElemNamesDict(elem: Element, methods: Dict[str, str]):
	methodName = elem.get("name", None)
	if methodName:
		methods[methodName] = methodName
		methods[methodName.lower()] = methodName
		methodId = elem.get("id", None)
		if methodId:
			methods[methodId] = methodName
			# methods[methodId.lower()] = methodName

def getMethodNamesDict(elem: Element) -> Dict[str, str]:
	methods = {} # type: Dict[str, Element]
	if getElemTag(elem) == "method":
		getMethodElemNamesDict(elem, methods)
		return methods
	for child in elem.getchildren():
		getMethodElemNamesDict(child, methods)
	return methods


def elemIsAction(elem: Element) -> bool:
	tag = getElemTag(elem)
	if tag == "method":
		return True
	return False

def parseInputValue(valueRaw: str, typ: str) -> Tuple[Any, Optional[str]]:
	"""
		returns (value, error)
	"""
	if typ == "xs:string":
		return valueRaw, None
	if typ == "xs:float":
		try:
			return float(valueRaw), None
		except ValueError:
			return None, "invalid float value %r" % valueRaw
	if typ == "xs:int":
		try:
			return int(valueRaw), None
		except ValueError:
			return None, "invalid int value %r" % valueRaw
	if typ == "xs:boolean":
		valueRaw = valueRaw.lower()
		if valueRaw in ("true", "t"):
			return True, None
		if valueRaw in ("false", "f"):
			return False, None
		return None, "invalid boolean value %r" % valueRaw
	return None, "unsupported type %r" % typ


class VirtualDir:
	def __init__(self, elem, pathRel: str, pathAbs: str, parent: Optional["VirtualDir"]):
		self.elem = elem
		self.pathRel = pathRel
		self.pathAbs = pathAbs
		self.parent = parent
		self.options, self.optionsMinimal = elemChildOptions(elem)


class CLI():
	def __init__(self, resources: Element) -> None:
		self._resources = resources
		self._root = VirtualDir(resources, "", "/", None)
		self.selectRoot()
		self._email = ""
		self._authToken = ""
		self._urlParamByValue = {}

	def init(self) -> Optional[str]:
		auth, err = getAuth()
		if err:
			return err
		self._email, self._authToken = auth
		if isfile(lastPathFile):
			with open(lastPathFile) as f:
				self.selectPathAbs(f.read().strip())

	def setVirtualDir(self, new_vdir: VirtualDir) -> None:
		self._cwd = new_vdir
		self._prompt = self._cwd.pathAbs + " > "

	def selectRoot(self) -> None:
		self.setVirtualDir(self._root)

	def selectVirtualDir(self, new_vdir: VirtualDir) -> bool:
		elem = new_vdir.elem
		pathAbs = new_vdir.pathAbs
		if elemIsAction(elem):
			data = {}

			# err = self.askUrlParams(new_vdir, data)
			# if err:
			# 	print("< ERROR:", err)
			# 	return True

			requestElems = elem.xpath("*[local-name() = 'request']")
			if requestElems:
				err = self.askJsonParams(requestElems[0], pathAbs, data)
				if err:
					print("< ERROR:", err)
					return True

			resData, err = self.sendRequest(elem, pathAbs, data)
			if err:
				print("< ERROR:", err)
				return True
			print("< Response:", dataToPrettyJson(resData))
			return True

		cur_vdir = self._cwd
		self.setVirtualDir(new_vdir)

		if len(new_vdir.optionsMinimal) == 1:
			childPath, childElem = list(new_vdir.optionsMinimal.items())[0]
			if elemIsAction(childElem):
				if self.selectPathRel(childPath):
					self.selectVirtualDir(cur_vdir)
					return True
		
		return True

	def selectParentDir(self) -> Optional[str]: # returns error
		if self._cwd.parent is None:
			return "no parent for %r" % self._cwd.pathAbs
		
		if not self.selectVirtualDir(self._cwd.parent):
			return "failed to switch to parent"

		return

	def selectPathAbs(self, pathAbs: str) -> True:
		if not pathAbs.startswith("/"):
			raise RuntimeError("selectPathAbs: invalid pathAbs=%r" % pathAbs)
		self.selectRoot()
		self.selectPathRel(pathAbs[1:])

	def selectPathRel(self, pathRel: str) -> True:
		# print("selectPathRel: %r" % pathRel)
		if pathRel.startswith("/"):
			raise RuntimeError("selectPathRel: invalid pathRel=%r" % pathRel)
		
		elem = self._cwd.options.get(pathRel, None)

		parts = pathRel.rstrip("/").split("/")
		if elem is None and len(parts) > 1:
			if self.selectPathRel(parts[0]+"/"):
				return self.selectPathRel("/".join(parts[1:])+"/")
			return False

		part = parts[0]
		if elem is None:
			elem = self._cwd.options.get(part, None)
		if elem is None:
			elem = self._cwd.options.get(part+"/", None)
		if elem is None:
			partName = self._urlParamByValue.get(part)
			if partName is not None:
				elem = self._cwd.options.get(partName, None)
		if elem is None:
			return False


		# FIXME: 
		if "{" in pathRel:
			parsedPath = parse_format(pathRel, pathRel)
			# example: parsedPath.named == {'var1': '{var1}', 'var2': '{var2}'}
			formatDict = {}
			for name in parsedPath.named:
				try:
					value = prompt(
						"> URL Parameter: " + name + " = ",
						history=FileHistory(self.paramHistoryPath(name)),
						auto_suggest=AutoSuggestFromHistory(),
						# completer=completer,
					)
				except KeyboardInterrupt:
					return False
				if value == "":
					print("ERROR: %s can not be empty" % name)
					return False
				formatDict[name] = value
				self._urlParamByValue[value] = name
			
			pathRelNew = pathRel.format(**formatDict)
			self._urlParamByValue[pathRelNew] = pathRel
			# print("pathRel=%r, pathRelNew=%r" % (pathRel, pathRelNew))
			pathRel = pathRelNew

		pathAbs = self._cwd.pathAbs + pathRel

		return self.selectVirtualDir(VirtualDir(elem, "", pathAbs, self._cwd))


	def selectPath(self, path: str) -> True:
		if path.startswith("/"):
			return self.selectPathAbs(path)
		return self.selectPathRel(path)

	def askJsonParams(self, requestElem: Element, path: str, data: Dict[str, Any]) -> Optional[str]:
		"""
			recursive function to ask all json/body parameters
			updates `data` argument
			returns error or None
		"""
		# FIXME: do we need the path
		for child in requestElem.getchildren():
			t = getElemTag(child)
			if t == "param":
				name = child.get("name", "")
				if not name:
					print("WARNING: element %r with tag %r has no name", child, t)
					continue
				typ = child.get("type", "")
				if not typ:
					print("WARNING: element %r with tag %r has no type", child, t)
					continue
				completer = getParamCompleter(child)
				valueRaw = prompt(
					"> Parameter: " + name + " = ",
					history=FileHistory(self.paramHistoryPath(name)),
					auto_suggest=AutoSuggestFromHistory(),
					completer=completer,
				)
				if valueRaw != "":
					value, err = parseInputValue(valueRaw, typ)
					if err:
						return err
					data[name] = value

	# returns (responseDict, error)
	# path argument ends with "/GET" or "/POST" or "/getUserInfo" for example
	# data is the dicty that is going to become request body (in json)
	def sendRequest(self, elem: Element, path: str, data: Dict[str, Any]) -> Tuple[Optional[Dict], str]:
		pathParts = path.split("/")
		methodsDict = getMethodNamesDict(elem)
		methodInput = pathParts[-1]
		method = methodsDict.get(methodInput, None)
		if not method:
			return None, "invalid method: " + methodInput + ", available: %s" % list(methodsDict.keys())
		url = baseURL + "/".join(pathParts[:-1])
		kwargs = {
			"headers": {"Authorization": "bearer " + self._authToken},
		}
		if data or method in ("PUT", "POST", "PATCH"):
			kwargs["json"] = data
			print("< Sending %s request to %s with json=%s" % (method, url, data))
		else:
			print("< Sending %s request to %s" % (method, url))
		try:
			res = requests.request(method, url, **kwargs)
		except Exception as e:
			return None, str(e)
		try:
			resData = res.json()
		except:
			return None, "non-json data: " + r.text
		err = ""
		if isinstance(resData, dict):
			err = resData.get("error", "")
		return resData, err

	def currentHistoryPath(self) -> str:
		pathAbs = self._cwd.pathAbs
		if pathAbs == "/":
			fname = "root"
		else:
			fname = pathAbs.strip("/").replace("/", "_")
		return join(histDir, fname)

	def paramHistoryPath(self, name: str) -> str:
		return join(histDir, "param-"+name)

	def runcmd(self, line) -> Optional[str]: # returns error
		if not line:
			return
		# print("runcmd:", line)
		line = line.strip()

		if line == "..":
			return self.selectParentDir()

		if self.selectPath(line):
			return
		
		if self.selectPath(line + "/"):
			return
		
		return "invalid option: " + line

	def finish(self):
		with open(lastPathFile, "w") as f:
			f.write(self._cwd.pathAbs)

	def cmdloop(self):
		while True:
			completer = WordCompleter(
				[key for key in self._cwd.options],
				ignore_case=False,
			)
			try:
				user_input = prompt(
					self._prompt,
					history=FileHistory(self.currentHistoryPath()),
					auto_suggest=AutoSuggestFromHistory(),
					completer=completer,
				)
			except KeyboardInterrupt:
				return
			
			err = self.runcmd(user_input)
			if err:
				print("< ERROR: %s" % err)


resources = doc.getchildren()[0]
assert(getElemTag(resources) == "resources")


cli = CLI(resources)
err = cli.init()
if err:
	raise Exception(err)
cli.cmdloop()
cli.finish()

# https://docs.python.org/3/library/cmd.html#cmd.Cmd.cmdloop