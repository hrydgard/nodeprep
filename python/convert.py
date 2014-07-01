from xmpp_stringprep import nodeprep
import json
import logging
import sys
import codecs
import os

sys.stdout = codecs.getwriter("utf-8")(sys.stdout)

for line in sys.stdin.read().split("\n")[:-1]:
	try:
		a = unicode(line.strip(), "utf-8")
		try:
			print nodeprep.prepare(a)
		except UnicodeError as e:
			print "ILLEGAL:unable to prepare:" + str(e)	
	except UnicodeError as e:
		print "ILLEGAL:unable to decode:" + str(e)
