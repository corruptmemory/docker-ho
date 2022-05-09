package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type templateData struct {
	Time string
}

const port = 8080
const homeText = `<!DOCTYPE html>
<html>
  <head>
    <title>Docker HO!</title>
  </head>
  <body>
  	<h1>Hello from the world of Docker!</h1>
  	<ul>
  	  <li>{{.Time}}</li>
  	</ul>
  </body>
</html>
`

const streamPageText = `<!DOCTYPE html>
<html>
  <head>
    <title>Streaming results example.</title>
  </head>
  <body>
  	<h1>Streaming results</h1>
  	<h3>Server time: {{.Time}}</h3>
  	<ul>
  	</ul>
	<script>
const list = document.querySelector('ul');
const decoder = new TextDecoder("utf-8");

// Fetch the original image
fetch('./stream-data')
// Retrieve its body as ReadableStream
.then(response => {
  const reader = response.body.getReader();
  return new ReadableStream({
    start(controller) {
      return pump();
      function pump() {
        return reader.read().then(({ done, value }) => {
          // When no more data needs to be consumed, close the stream
          if (done) {
            controller.close();
            let listItem = document.createElement('li');
            listItem.textContent = "Done!";
            list.appendChild(listItem);
            return;
          }
          var decoded = decoder.decode(value);
		  let listItem = document.createElement('li');
		  listItem.textContent = decoded;
		  list.appendChild(listItem);
          return pump();
        });
      }
    }
  })
})
.catch(err => console.error(err));
	</script>
  </body>
</html>
`

var home = template.Must(template.New("home").Parse(homeText))
var stream = template.Must(template.New("stream").Parse(streamPageText))

func homeHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	t := time.Now()
	ts := t.Format(time.RFC1123Z)
	td := templateData{Time: ts}
	home.Execute(w, &td)
}

func streamPageHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	t := time.Now()
	ts := t.Format(time.RFC1123Z)
	td := templateData{Time: ts}
	stream.Execute(w, &td)
}

func streamDataHandler(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(w, "Chunk #%d\n", i)
		flusher.Flush() // Trigger "chunked" encoding and send a chunk...
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	log.Printf("Starting web server on port %d", port)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/stream", streamPageHandler)
	http.HandleFunc("/stream-data", streamDataHandler)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
