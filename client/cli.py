#!/usr/bin/python3

import json
import os
import sys
from datetime import datetime, timedelta
from os.path import abspath, dirname, isfile, join

import dateutil.tz
import requests
from dateutil.parser import parse as parse_time
from dateutil.tz.tz import tzutc

try:
	from parse import parse as parse_format
except ModuleNotFoundError as e:
	e.msg += ", run `sudo pip3 install parse` to install"
	raise e


from typing import Any, Optional

from lxml import etree
from lxml.etree import _Element as Element

# the code for cmd.Cmd is very ugly and hard to understan
# readline's complete func silently (and stupidly) hides any exception
# and only shows the print if it's in the first line of function. very awkward!
# import atexit
from prompt_toolkit import prompt as promptLow
from prompt_toolkit.auto_suggest import AutoSuggestFromHistory
from prompt_toolkit.completion.word_completer import WordCompleter
from prompt_toolkit.history import FileHistory

utc = tzutc()


host = os.getenv("STARCAL_HOST")
if not host:
	import socket

	host = socket.gethostname()

secure = os.getenv("STARCAL_HOST_SECURE") == "1"

port = "9001"

protocol = "https" if secure else "http"
baseURL = f"{protocol}://{host}:{port}"

print("< Base URL:", baseURL)


myDir = dirname(__file__)
cwd = os.getcwd()
if myDir in {".", ""}:
	myDir = cwd
elif os.sep == "/":
	if myDir.startswith("./"):
		myDir = cwd + myDir[1:]
	elif myDir[0] != "/":
		myDir = join(cwd, myDir)
elif os.sep == "\\" and myDir.startswith(".\\"):
	myDir = cwd + myDir[1:]
# print("myDir={myDir!r}")

rootDir = abspath(dirname(myDir))
docPath = join(rootDir, "docs", "api-v1.wadl")

homeDir = os.getenv("HOME")
confDir = join(homeDir, ".starcal-server", "cli")
histDir = join(confDir, "history")
tokenDirParent = join(confDir, "auth-tokens")
tokenDir = join(tokenDirParent, host)
lastPathFile = join(confDir, "last-path")

for direc in (confDir, histDir, tokenDirParent, tokenDir):
	os.makedirs(direc, mode=0o700, exist_ok=True)

indent = "\t"


pathGeneralizeDict = {
	"event/allDayTask": "event/{eventType}",
	"event/custom": "event/{eventType}",
	"event/dailyNote": "event/{eventType}",
	"event/largeScale": "event/{eventType}",
	"event/lifetime": "event/{eventType}",
	"event/monthly": "event/{eventType}",
	"event/task": "event/{eventType}",
	"event/universityClass": "event/{eventType}",
	"event/universityExam": "event/{eventType}",
	"event/weekly": "event/{eventType}",
	"event/yearly": "event/{eventType}",
}
pathGeneralizeIdByPath = {
	"event/{eventType}": "event_eventType_resource",
}


with open(docPath, encoding="utf-8") as docFile:
	doc = etree.XML(docFile.read().encode("utf-8"))


def dataToPrettyJson(data, ensure_ascii=False, sort_keys=False):
	return json.dumps(
		data,
		sort_keys=sort_keys,
		indent=2,
		ensure_ascii=ensure_ascii,
	)


def prompt(
	message: str,
	multiline: bool = False,
	**kwargs,
):
	text = promptLow(message=message, **kwargs)
	if multiline and text == "!m":
		print("Entering Multi-line mode, press Alt+Enter to end")
		text = promptLow(message="", multiline=True, **kwargs)
	return text


def paramHistoryPath(name: str) -> str:
	return join(histDir, f"param-{name}")


def getEmail() -> str:
	email = os.getenv("STARCAL_EMAIL")
	if email:
		return email
	while True:
		try:
			email = prompt(
				"Email: ",
				history=FileHistory(join(histDir, "email")),
				auto_suggest=AutoSuggestFromHistory(),
			)
		except KeyboardInterrupt:
			sys.exit(0)
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
def getSavedToken(email: str) -> str | None:
	tokenPath = join(tokenDir, email)
	if not isfile(tokenPath):
		print("Token file does not exist:", tokenPath)
		return None
	with open(tokenPath, encoding="utf-8") as tokenFile:
		tokenJson = tokenFile.read()
	tokenDict = json.loads(tokenJson)
	# print(tokenDict)
	expStr = tokenDict.get("expiration", "")
	if not expStr:
		print(f"WARNING: invalid token file {tokenPath!r}, no 'expiration'")
		return None
	token = tokenDict.get("token", "")
	if not token:
		print(f"WARNING: invalid token file {tokenPath!r}, no 'token'")
		return None
	exp = parse_time(expStr)
	if exp - timedelta(minutes=5) < datetime.now(tz=localTZ):
		print(
			f"Saved token in {tokenPath} is expired"
			f" or about to be expired on {exp}"
		)
		return None
	return token


def deleteSavedToken(email: str) -> bool:
	tokenPath = join(tokenDir, email)
	if not isfile(tokenPath):
		return False
	os.remove(tokenPath)
	print(f"removed token file {tokenPath}")
	return True


# returns (email, token), error
def getAuth() -> tuple[tuple[str, str] | None, str | None]:
	email = getEmail()
	token = getSavedToken(email)
	if token:
		return (email, token), None
	password = getPassword()
	if not password:
		return None, "password is empty"
	url = baseURL + "/auth/login/"
	print("Sending login request")
	res = requests.post(
		url,
		json={
			"email": email,
			"password": password,
		},
	)
	err = None
	try:
		resData = res.json()
	except Exception:  # simplejson.errors.JSONDecodeError
		resData = None
	else:
		err = resData.get("error")
	if res.status_code != 200:
		if not err:
			err = f"{res} from {url}"
		return None, err
	if resData is None:
		return "", f"ERROR: non-json body from {url!r}"
	token = resData.get("token")
	if not token:
		return None, "login returned no token"

	try:
		expStr = resData["expiration"]
	except KeyError:
		return None, "login returned no expiration"

	tokenPath = join(tokenDir, email)
	with open(tokenPath, "w", encoding="utf-8") as tokenFile:
		json.dump(
			{
				"token": token,
				"expiration": expStr,
			},
			tokenFile,
		)

	return (email, token), None


def getElemTag(elem) -> str:
	# for example, elem.tag == "{http://localhost:9001/application.wadl}resource"
	return elem.tag.split("}")[1]


def elemName(elem) -> str:
	return elem.get("name", "NO_NAME")


def elemID(elem) -> str:
	return elem.get("id", "NO_ID")


def elemType(elem) -> str:
	return elem.get("type", "NO_TYPE")


def elemPath(elem) -> str:
	return elem.get("path", "NO_PATH")


def elemValue(elem) -> str:
	return elem.get("value", "NO_VALUE")


def elemRepr(elem) -> str:
	tag = getElemTag(elem)
	if tag == "resource":
		return elemPath(elem)
	if tag == "method":
		return elemName(elem) + f" ({elemID(elem)})"
	if tag in {"param", "element"}:
		return elemName(elem) + f" (type={elemID(elem)})"
	if tag == "item":
		return f"(type={elemType(elem)})"
	if tag == "option":
		return elemValue(elem)
	if tag == "representation":
		return ""
	return tag


def printElem(elem: Element, level: int):
	tag = getElemTag(elem)
	prefix = indent * level + getElemTag(elem)
	if tag == "resource":
		print(f"{prefix}: {elemPath(elem)}")
	elif tag == "method":
		print(f"{prefix}: {elemName(elem)} ({elemID(elem)})")
	elif tag in {"param", "element"}:
		print(f"{prefix}: {elemName(elem)} (type={elemType(elem)})")
	elif tag == "item":
		print(f"{prefix} (type={elemType(elem)})")
	elif tag == "option":
		print(f"{prefix}: {elemValue(elem)}")
	elif tag == "representation":
		pass
	else:
		print(prefix)
	for child in elem.getchildren():
		printElem(child, level + 1)


def nonEmptyStrings(*args) -> list[str]:
	return [
		x
		for x in args
		if x
	]


def elemKeys(elem, parentElem) -> str:
	# returns virtual file names of element
	tag = getElemTag(elem)
	if tag == "resource":
		return nonEmptyStrings(elem.get("path", None))
	if tag in {"method", "element"}:
		return nonEmptyStrings(elem.get("name", None), elem.get("id", None))
	if tag == "param":
		if getElemTag(parentElem) == "resource":
			return []
		return nonEmptyStrings(elem.get("name", None), elem.get("id", None))
	# if tag == "item":
	# 	return [] # FIXME
	# if tag == "option":
	# 	return nonEmptyStrings(elem.get("value", None))
	if tag == "representation":
		return []

	return []
	# print("elemPath", prefix)


# returns (options, optionsMinimal)
def elemChildOptions(
	elem: Element,
) -> tuple[
	dict[str, Element],
	dict[str, Element],
]:
	options = {}  # type: Dict[str, Element]
	optionsMinimal = {}  # type: Dict[str, Element]
	for child in elem.getchildren():
		keys = elemKeys(child, elem)
		if not keys:
			continue
		optionsMinimal[keys[0]] = child
		for key in keys:
			if key.strip("/") in pathGeneralizeIdByPath:
				continue
			options[key] = child
			options[key.lower()] = child
	return options, optionsMinimal


def getParamCompleter(elem: Element) -> WordCompleter | None:
	options = []  # type: List[str]
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


def getMethodElemNamesDict(elem: Element, methods: dict[str, str]):
	methodName = elem.get("name", None)
	if methodName:
		methods[methodName] = methodName
		methods[methodName.lower()] = methodName
		methodId = elem.get("id", None)
		if methodId:
			methods[methodId] = methodName
			# methods[methodId.lower()] = methodName


def getMethodNamesDict(elem: Element) -> dict[str, str]:
	methods = {}  # type: Dict[str, Element]
	if getElemTag(elem) == "method":
		getMethodElemNamesDict(elem, methods)
		return methods
	for child in elem.getchildren():
		getMethodElemNamesDict(child, methods)
	return methods


def elemIsAction(elem: Element) -> bool:
	tag = getElemTag(elem)
	return tag == "method"


def getChildrenWithTag(elem: Element, tag: str) -> Element:
	return [child for child in elem.getchildren() if getElemTag(child) == tag]


def parseInputValue(valueRaw: str, typ: str) -> tuple[Any, str | None]:
	"""Returns (value, error)."""
	if typ == "xs:string":
		return valueRaw, None
	if typ == "xs:float":
		try:
			return float(valueRaw), None
		except ValueError:
			return None, f"invalid float value {valueRaw!r}"
	if typ == "xs:int":
		try:
			return int(valueRaw), None
		except ValueError:
			return None, f"invalid int value {valueRaw!r}"
	if typ == "xs:boolean":
		valueRaw = valueRaw.lower()
		if valueRaw in {"true", "t"}:
			return True, None
		if valueRaw in {"false", "f"}:
			return False, None
		return None, f"invalid boolean value {valueRaw!r}"
	return None, f"unsupported type {typ!r}"


def updateOptionsDict(
	options: dict[str, list[Element]],
	elemEptions: dict[str, Element],
):
	for key, elem in elemEptions.items():
		if key in options:
			options[key].append(elem)
		else:
			options[key] = [elem]


class VirtualDir:
	def __init__(
		self,
		elems: list[Element],
		pathRel: str,
		pathAbs: str,
		parent: Optional["VirtualDir"],
	) -> None:
		self.elems = elems
		self.pathRel = pathRel
		self.pathAbs = pathAbs
		self.parent = parent
		options = {}
		optionsMinimal = {}
		for elem in elems:
			elemEptions, elemOptionsMinimal = elemChildOptions(elem)
			updateOptionsDict(options, elemEptions)
			updateOptionsDict(optionsMinimal, elemOptionsMinimal)
		self.options = options  # type: Dict[str, List[Element]]
		self.optionsMinimal = optionsMinimal  # type: Dict[str, List[Element]]


class CLI:
	def __init__(self, resources: Element) -> None:
		self._resources = resources
		self._root = VirtualDir([resources], "", "/", None)
		self.selectRoot()
		self._email = ""
		self._authToken = ""
		self._urlParamByValue = {}

	def init(self) -> str | None:
		auth, err = getAuth()
		if err:
			return err
		self._email, self._authToken = auth
		if isfile(lastPathFile):
			with open(lastPathFile, encoding="utf-8") as f:
				self.selectPathAbs(f.read().strip())
				return None
		return None

	def setVirtualDir(self, new_vdir: VirtualDir) -> None:
		self._cwd = new_vdir
		self._prompt = self._cwd.pathAbs + " > "

	def selectRoot(self) -> None:
		self.setVirtualDir(self._root)

	def selectVirtualDir(self, new_vdir: VirtualDir) -> bool:
		elems = new_vdir.elems
		pathAbs = new_vdir.pathAbs
		if len(elems) == 1 and elemIsAction(elems[0]):
			elem = elems[0]
			data = {}

			# err = self.askUrlParams(new_vdir, data)
			# if err:
			# 	print("< ERROR:", err)
			# 	return True

			requestElems = elem.xpath("*[local-name() = 'request']")
			if requestElems:
				err = self.askJsonParams(requestElems[0], data)
				if err:
					print("< ERROR:", err)
					return True

			resData, err = self.sendRequest(elem, pathAbs, data)
			if err:
				print("< ERROR:", err)
				return True
			print("< Response:", dataToPrettyJson(resData))
			return True

		if not pathAbs.endswith("/"):
			pathAbs += "/"
			new_vdir.pathAbs = pathAbs

		cur_vdir = self._cwd
		self.setVirtualDir(new_vdir)

		if len(new_vdir.optionsMinimal) == 1:
			childPath, childElems = next(iter(new_vdir.optionsMinimal.items()))
			if len(childElems) == 1 and elemIsAction(childElems[0]):  # noqa: SIM102
				if self.selectPathRel(childPath):
					self.selectVirtualDir(cur_vdir)
					return True

		return True

	def selectParentDir(self) -> str | None:  # returns error
		if self._cwd.parent is None:
			return f"no parent for {self._cwd.pathAbs!r}"

		if not self.selectVirtualDir(self._cwd.parent):
			return "failed to switch to parent"

		return None

	def selectPathAbs(self, pathAbs: str) -> True:
		if not pathAbs.startswith("/"):
			raise RuntimeError(f"selectPathAbs: invalid pathAbs={pathAbs!r}")
		self.selectRoot()
		self.selectPathRel(pathAbs[1:])

	def selectPathRel(self, pathRel: str) -> True:  # noqa: C901, PLR0912
		# print(f"selectPathRel: {pathRel!r}")
		if pathRel.startswith("/"):
			raise RuntimeError(f"selectPathRel: invalid pathRel={pathRel!r}")

		elems = self._cwd.options.get(pathRel, [])

		if len(elems) > 1:
			pass  # FIXME

		parts = pathRel.rstrip("/").split("/")
		if not elems and len(parts) > 1:
			if self.selectPathRel(parts[0] + "/"):
				return self.selectPathRel("/".join(parts[1:]) + "/")
			return False

		part = parts[0]
		if not elems:
			elems = self._cwd.options.get(part, [])
		if not elems:
			elems = self._cwd.options.get(part + "/", [])
		if not elems:
			partName = self._urlParamByValue.get(part)
			if partName is not None:
				elems = self._cwd.options.get(partName, [])
		if not elems:
			return False

		# FIXME:
		if "{" in pathRel:
			parsedPath = parse_format(pathRel, pathRel)
			# example: parsedPath.named == {'var1': '{var1}', 'var2': '{var2}'}
			formatDict = {}
			for name in parsedPath.named:
				try:
					value = prompt(
						f"> URL Parameter: {name} = ",
						history=FileHistory(paramHistoryPath(name)),
						auto_suggest=AutoSuggestFromHistory(),
						# completer=completer,
					)
				except KeyboardInterrupt:
					return False
				if not value:
					print(f"ERROR: {name} can not be empty")
					return False
				formatDict[name] = value
				self._urlParamByValue[value] = name

			pathRelNew = pathRel.format(**formatDict)
			self._urlParamByValue[pathRelNew] = pathRel
			# print(f"pathRel={pathRel!r}, pathRelNew={pathRelNew!r}")
			pathRel = pathRelNew

		pathAbs = self._cwd.pathAbs + pathRel

		vdirElems = elems
		secondElemPath = pathGeneralizeDict.get(pathAbs.strip("/"))
		if secondElemPath:
			secondElemId = pathGeneralizeIdByPath[secondElemPath]
			tmpElems = self._resources.xpath(f"//*[@id='{secondElemId}']")
			if tmpElems:
				if len(tmpElems) > 1:
					print(
						f"Error: {len(tmpElems)} elements found "
						f"with id='{secondElemId}'"
					)
				vdirElems += tmpElems
			else:
				print(f"Error: No element found with id='{secondElemId}'")

		vdir = VirtualDir(vdirElems, "", pathAbs, self._cwd)
		return self.selectVirtualDir(vdir)

	def selectPath(self, path: str) -> True:
		if path.startswith("/"):
			return self.selectPathAbs(path)
		return self.selectPathRel(path)

	@staticmethod
	def askJsonParams(
		requestElem: Element,
		data: dict[str, Any],
	) -> str | None:
		"""
		recursive function to ask all json/body parameters
		updates `data` argument
		returns error or None.
		"""
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
				multiline = child.get("multiline", "") == "true"
				history = None
				if child.get("secret", "") != "true":
					history = FileHistory(paramHistoryPath(name))
				try:
					valueRaw = prompt(
						f"> Parameter: {name} = ",
						multiline=multiline,
						history=history,
						auto_suggest=AutoSuggestFromHistory(),
						completer=completer,
					)
				except KeyboardInterrupt:
					return "Canceled"
				if not valueRaw:
					value, err = parseInputValue(valueRaw, typ)
					if err:
						return err
					data[name] = value
		return None

	# returns (responseDict, error)
	# path argument ends with "/GET" or "/POST" or "/getUserInfo" for example
	# data is the dicty that is going to become request body (in json)
	def sendRequest(
		self,
		elem: Element,
		path: str,
		data: dict[str, Any],
	) -> tuple[dict | None, str]:
		pathParts = path.split("/")
		methodsDict = getMethodNamesDict(elem)
		methodInput = pathParts[-1]
		method = methodsDict.get(methodInput, None)
		if not method:
			return None, (
				f"invalid method: {methodInput}, "
				f"available: {list(methodsDict.keys())}"
			)
		url = baseURL + "/".join(pathParts[:-1])
		kwargs = {
			"headers": {"Authorization": "bearer " + self._authToken},
		}
		if method in {"PUT", "POST", "PATCH"}:
			kwargs["json"] = data
			print(f"< Sending {method} request to {url} with json={data}")
		elif data:
			params = list(data.items())
			kwargs["params"] = params
			print(f"< Sending {method} request to {url} with params={params}")
		else:
			print(f"< Sending {method} request to {url}")
		try:
			res = requests.request(method, url, **kwargs)
		except Exception as e:
			return None, str(e)
		try:
			resData = res.json()
		except Exception:
			return None, f"non-json data: {res.text}"
		err = ""
		if isinstance(resData, dict):
			err = resData.get("error", "")
		if not err and path == "/auth/logout/POST":
			deleteSavedToken(self._email)
			# FIXME: should we clear self._authToken and ask user to login again?

		return resData, err

	def currentHistoryPath(self) -> str:
		pathAbs = self._cwd.pathAbs
		fname = "root" if pathAbs == "/" else pathAbs.strip("/").replace("/", "_")
		return join(histDir, fname)

	def runcmd(self, line) -> str | None:  # returns error
		if not line:
			return None
		# print("runcmd:", line)
		line = line.strip()

		if line == "..":
			return self.selectParentDir()

		if self.selectPath(line):
			return None

		if self.selectPath(line + "/"):
			return None

		return f"invalid option: {line}"

	def finish(self):
		with open(lastPathFile, "w", encoding="utf-8") as f:
			f.write(self._cwd.pathAbs)

	def cmdloop(self):
		while True:
			completer = WordCompleter(
				list(self._cwd.options),
				ignore_case=False,
			)
			try:
				user_input = prompt(
					self._prompt,
					history=FileHistory(self.currentHistoryPath()),
					auto_suggest=AutoSuggestFromHistory(),
					completer=completer,
				)
			except (KeyboardInterrupt, EOFError):
				return

			err = self.runcmd(user_input)
			if err:
				print(f"< ERROR: {err}")


resources = doc.getchildren()[0]
assert getElemTag(resources) == "resources"


cli = CLI(resources)
err = cli.init()
if err:
	raise Exception(err)
cli.cmdloop()
cli.finish()

# https://docs.python.org/3/library/cmd.html#cmd.Cmd.cmdloop
