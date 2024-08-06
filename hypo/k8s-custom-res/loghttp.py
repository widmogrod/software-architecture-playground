from http.server import BaseHTTPRequestHandler, HTTPServer

class RequestHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"{self.client_address[0]} - - [{self.log_date_time_string()}] {format % args}")

    def do_GET(self):
        self.log_message('"%s" %s', self.requestline, str(self.headers))
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b'Hello, world!')

    def do_POST(self):
        self.log_message('"%s" %s', self.requestline, str(self.headers))
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        self.log_message('Body: %s', post_data.decode('utf-8'))
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b'Received POST request')

def run(server_class=HTTPServer, handler_class=RequestHandler, port=8080):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f'Starting httpd server on port {port}')
    httpd.serve_forever()

if __name__ == "__main__":
    run()