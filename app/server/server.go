package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gaurav-nagpal-new/go-http-server-scratch/constant"
)

// HTTPRequest struct to store fields to store fields related to request
type HTTPRequest struct {
	HTTPMethod  string            // defines the http method like GET POST
	URL         string            // define the URL at which request is reached
	Headers     map[string]string // headers from the requst
	RequestBody []byte
	Response    []byte
}

// parseReq() method to parse above fields and initialize the struct and use that - define methods on this.
func (h *HTTPRequest) parseReq(conn net.Conn) {
	switch {
	case h.URL == "/":
		h.Response = fmt.Appendf(h.Response, "HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s%s",
			constant.CRLF, constant.CRLF, len(h.RequestBody), constant.CRLF, constant.CRLF, h.RequestBody)
	case strings.HasPrefix(h.URL, "/echo"):
		// request contains some body in the request url, send that in response
		pathParts := strings.Split(h.URL, "/")
		body := pathParts[2]

		h.Response = fmt.Appendf(h.Response, "HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s%s", constant.CRLF, constant.CRLF, len(body), constant.CRLF, constant.CRLF, body)
	case h.URL == "/user-agent":
		// read header from request and send that in response, headers start from 3 index and goes till end - 1, skips body in the last

		/*
			request body:
			// Request line
			GET
			/user-agent
			HTTP/1.1
			\r\n

			// Headers
			Host: localhost:4221\r\n
			User-Agent: foobar/1.2.3\r\n  // Read this value
			Accept: */ //*\r\n
		// \r\n

		// Request body (empty)

		body := h.Headers["User-Agent"]
		contentLength := utf8.RuneCountInString(body)
		h.Response = fmt.Appendf(h.Response, "HTTP/1.1 200 OK%sContent-Type: text/plain%sContent-Length: %d%s%s%s",
			constant.CRLF, constant.CRLF, contentLength, constant.CRLF, constant.CRLF, body)

	case strings.HasPrefix(h.URL, "/files"):
		pathParts := strings.Split(h.URL, "/")
		fileName := pathParts[2]
		filePath := fmt.Sprintf("../temp/%s.txt", fileName)

		// if its a GET request, send the file content
		if strings.EqualFold(http.MethodGet, h.HTTPMethod) {
			// check if file exists or not, if it does not already exist, then return 404
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				fmt.Appendf(h.Response, "HTTP/1.1 404 Not Found%sContent-Type: text/plain%sContent-Length: %d%s%s%s",
					constant.CRLF, constant.CRLF, len(h.RequestBody), constant.CRLF, constant.CRLF, h.RequestBody)
			}

			// otherwise, send the response with file content
			content, _ := os.ReadFile(filePath)
			contentLength := utf8.RuneCountInString(string(content))
			h.Response = fmt.Appendf(h.Response, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", contentLength, content)
		} else if strings.EqualFold(http.MethodPost, h.HTTPMethod) {
			// if its a POST request, then create a new file and write the content received in request body
			// read Content-Length to know how many characters to read from the request body

			contentLength, _ := strconv.Atoi(h.Headers[constant.ContentLengthHeader])

			// now read these characters from the request body
			requestBody := h.RequestBody

			// check if we need to do the compression
			if encodingScheme, ok := h.Headers[constant.AcceptEncodingHeader]; ok {
				// currently only support gzip by default
				encodedContent := new(bytes.Buffer)
				gz := gzip.NewWriter(encodedContent)

				content := []byte("your raw content here")

				if _, err := gz.Write(content); err != nil {
					log.Fatal("error writing content:", err)
				}
				gz.Close()

				contentLength := len(encodedContent.Bytes())

				h.Response = fmt.Appendf(h.Response,
					"HTTP/1.1 200 OK%sContent-Encoding: %s%sContent-Length: %d%s%s%s",
					constant.CRLF, encodingScheme, constant.CRLF, contentLength, constant.CRLF, constant.CRLF, encodedContent.Bytes(),
				)
			}

			// Only read according to Content-Length header
			requestBody = requestBody[:contentLength]

			// write the body to the new file
			os.WriteFile(filePath, []byte(requestBody), 0666)
		}
	default:
		// response:
		/*
			HTTP/1.1 404 Not Found\r\n
			Content-Type: text/plain\r\n
			Content-Length: 6\r\n
			\r\n
			404 page not found
		*/
		body := "404 page not found"
		h.Response = fmt.Appendf(h.Response, "HTTP/1.1 404 Not Found%sContent-Type: text/plain%sContent-Length: %d%s%s%s",
			constant.CRLF, constant.CRLF, len(body), constant.CRLF, constant.CRLF, body)
	}

	h.sendResponse(conn)
}

func initializeHTTPRequest(request []byte) *HTTPRequest {

	requestParts := strings.Split(string(request), constant.CRLF)
	/*
		part 1 represents the request line
		part 2 represents the headers -> but it breaks on constant.CRLF so from part 2 it contains headers
		part 3 represents the optional request body
	*/

	requestLine := strings.Split(requestParts[0], " ")
	// requestLine[0] represents the http request method
	// requestLine[1] represents the http request URL

	headers := fetchHeadersFromRequest(requestParts)
	url := requestLine[1]
	requestBody := requestParts[len(requestParts)-1]
	contentLength, _ := strconv.Atoi(headers[constant.ContentLengthHeader])

	return &HTTPRequest{
		HTTPMethod:  requestLine[0],
		URL:         url,
		Headers:     headers,
		RequestBody: []byte(requestBody[:contentLength]),
	}
}

func fetchHeadersFromRequest(requestParts []string) map[string]string {

	headersMap := map[string]string{}
	for i := 3; i < len(requestParts)-2; i++ {
		singleHeader := strings.Split(requestParts[i], ": ")
		headersMap[singleHeader[0]] = strings.Join(singleHeader[1:], " ")
	}

	return headersMap
}

func (h *HTTPRequest) sendResponse(con net.Conn) {
	con.Write((h.Response))
}

func HandleConnection(con net.Conn) {

	defer con.Close()
	// send alternate response based on the URL.
	request := make([]byte, 1024)
	con.Read(request) // store the request body in request variable

	httpRequest := initializeHTTPRequest(request)
	httpRequest.parseReq(con)
}

func startHttpServer(network, host string) error {
	// first bind a tcp server to a port
	// 0.0.0.0 -> It means the server will accept incoming connections from anywhere, on any of the host’s network interfaces.
	l, err := net.Listen(network, host)
	if err != nil {
		return err
	}

	// now accept connections
	/*
		The method l.Accept():

		Waits (blocks) until a client connects to your server.

		When a client connects, it:

		Returns a new socket (net.Conn) specifically for communication with that client.

		Leaves the original l (the server socket) still listening for new connections as there can be more clients waiting for the listen
	*/

	fmt.Println("Starting server....")

	// accept infinite number of concurrent connections
	for {
		con, err := l.Accept()
		if err != nil {
			return err
		}
		// defer does not run this in an infinite loop as it only runs when surrounding function exits and infinite loop does not exit ever
		go HandleConnection(con)
	}
}

func main() {

	// start the server
	const network = "tcp"
	const host = "0.0.0.0:8080"
	if err := startHttpServer(network, host); err != nil {
		log.Fatal("error starting http server", err.Error())
	}

}
