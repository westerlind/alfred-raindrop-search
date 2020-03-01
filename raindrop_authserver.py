# Small web server for handling OAuth redirections from Raindrop.io,
# and thereby making OAuth authentication for a local script possible.
# Compatible with both Python 2.7 and 3.X to be somewhat future proof.
# 
# By Andreas Westerlind in 2020
# 

try:
  # Python 3
  from socketserver import TCPServer as HTTPServer
  from http.server import SimpleHTTPRequestHandler
  from urllib.parse import urlparse
  from urllib.parse import parse_qs
except ImportError:
  # Python 2
  from BaseHTTPServer import HTTPServer
  from SimpleHTTPServer import SimpleHTTPRequestHandler
  from urlparse import urlparse
  from urlparse import parse_qs
# Python 2 & 3
import os
import json

# Address and port to use for the webserver. Bind it to 127.0.0.1 so it's not externally reachable,
# and so that we don't get any uneccesary firewall allow/deny dialog
bind_address = ("127.0.0.1", 11038)

# Use this to kill the webserver, as it seems to be wierdly hard to reliably tell it to stop in a better way.
kill_command = "kill $(ps aux|grep raindrop_authserver.py|grep -v 'grep'|awk '{print $2}') 2> /dev/null"

# Define how to handle requests
class AuthRedirectionHandler(SimpleHTTPRequestHandler):
  def do_GET(self):
    try:
      # Get code from Raindrop
      code = parse_qs(urlparse(self.path).query).get("code", str)[0]
    except TypeError:
      # If a request was made without a code, just display an error message
      self.path = "/auth_error.html"
    else:
      # We got the code, so we use that to request a token from Raindrop, which we use a PHP script to do
      # because of issues with SSL root certificates in some versions of Python on macOS,
      # which makes this unreliable to do in Python if we should keep this thing small and simple but still safe.
      token_result = os.popen("/usr/bin/php raindrop-get-token.php '" + code.replace("'", "'\\''") + "'").read()
      if token_result == "success":
        # If we succeeded in getting the token, display information about the authentication being successful
        self.path = "/auth_info.html"
      else:
        # If we failed to get the token, display information about the authentication having failed
        self.path = "/auth_error.html"
    # Stop the webserver in 2 seconds, so we have time to display the info web page before it quits
    os.system("sleep 2 && " + kill_command + " &")
    return SimpleHTTPRequestHandler.do_GET(self)

# Start a countdown for killing the webserver after 20 minutes, so we don't accidentally leave it running forever
os.system("sleep 1200 && " + kill_command + " &")

# Start the web server
server = HTTPServer(bind_address, AuthRedirectionHandler)
server.serve_forever()