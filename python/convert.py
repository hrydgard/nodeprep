from xmpp_stringprep import nodeprep
import json
import logging
import sys
import codecs

sys.stdout = codecs.getwriter("utf-8")(sys.stdout)

for line in sys.stdin.read().split("\n")[:-1]:
	try:
		sys.stdout.write(nodeprep.prepare(unicode(line.strip(), "utf-8")))
		sys.stdout.write("\n")
	except UnicodeError as e:
		print "ILLEGAL"
	
