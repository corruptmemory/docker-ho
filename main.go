package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
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
          var parts = decoded.split("\n");
          for (var j = 0; j < parts.length; j++) {
            var p = parts[j];
            if (p.length > 0) {
              let listItem = document.createElement('li');
              listItem.textContent = p;
              list.appendChild(listItem);
            }
          }
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

type jobLogEntry struct {
	Done      bool
	Timestamp time.Time
	Entry     string
}

type jobHistory struct {
	started time.Time
	history []jobLogEntry
}

type jobDoerCommandType int

const (
	commandSubscribe jobDoerCommandType = iota
	commandUnsubscribe
)

type subscription struct {
	id     string
	doer   *jobDoerThinger
	events chan any
}

func (s *subscription) unsubscribe() {
	s.doer.commandSink <- commandMessage{
		command: commandUnsubscribe,
		id:      s.id,
	}
	s.id = ""
	s.doer = nil
	s.events = nil
}

type commandMessage struct {
	command jobDoerCommandType
	id      string
	result  chan subscription
}

type jobDoerThinger struct {
	history     jobHistory
	logSink     chan jobLogEntry
	commandSink chan commandMessage
}

func (j *jobDoerThinger) subscribe() subscription {
	r := make(chan subscription)
	j.commandSink <- commandMessage{
		command: commandSubscribe,
		result:  r,
	}
	return <-r
}

func (j *jobDoerThinger) log(in string) {
	ts := time.Now()
	j.logSink <- jobLogEntry{
		Timestamp: ts,
		Entry:     in,
	}
}

func (j *jobDoerThinger) done() {
	ts := time.Now()
	j.logSink <- jobLogEntry{
		Timestamp: ts,
		Done:      true,
	}
}

func newJobDoerThinger() (result *jobDoerThinger) {
	result = &jobDoerThinger{
		logSink:     make(chan jobLogEntry, 100),
		commandSink: make(chan commandMessage, 100),
	}
	go result.run()
	return
}

func (j *jobDoerThinger) run() {
	subscribers := map[string]subscription{}

	history := jobHistory{}

	newSubscriber := func(result chan subscription) {
		id := fmt.Sprintf("id-%v-%d", time.Now(), len(subscribers))
		s := subscription{
			id:     id,
			doer:   j,
			events: make(chan any, 100),
		}
		if len(history.history) > 0 {
			s.events <- history
		}
		if j.logSink == nil {
			close(s.events)
		}
		subscribers[id] = s
		result <- s
		close(result)
	}

	unsubscribe := func(id string) {
		tryClose := func(events chan any) {
			defer func() {
				if recover() != nil {
					// ignore
				}
			}()
			close(events)
		}

		if s, ok := subscribers[id]; ok {
			tryClose(s.events)
			s.doer = nil
			s.events = nil
			s.id = ""
			delete(subscribers, id)
		}
	}

	doLog := func(e jobLogEntry) {
		if len(history.history) == 0 {
			history.started = e.Timestamp
		}

		if !e.Done {
			history.history = append(history.history, e)
			for _, s := range subscribers {
				s.events <- e
			}
		} else {
			close(j.logSink)
			j.logSink = nil
			for _, s := range subscribers {
				close(s.events)
				s.events = nil
			}
		}
	}

	for {
		select {
		case l := <-j.logSink:
			doLog(l)
		case c := <-j.commandSink:
			switch c.command {
			case commandSubscribe:
				newSubscriber(c.result)
			case commandUnsubscribe:
				unsubscribe(c.id)
			}
		}
	}
}

var logger = newJobDoerThinger()

var startage = make(chan struct{})

func tehJob() {
	step := func(msg string, delay time.Duration) {
		logger.log(msg)
		log.Print(msg)
		time.Sleep(delay)
	}

	<-startage
	log.Print("Running!")

	step("Provisioning", 5*time.Second)
	step("Starting...", 5*time.Second)
	step("Doing stuff 1...", 2*time.Second)
	step("Doing stuff 2...", 3*time.Second)
	step("Doing stuff 3...", 1*time.Second)
	step("Doing stuff 4...", 5*time.Second)
	step("Doing stuff 5...", 6*time.Second)
	step("Doing stuff 6...", 10*time.Second)
	step("Doing stuff 7...", 1*time.Second)
	step("Doing stuff 8...", 2*time.Second)
	step("Doing stuff 9...", 1*time.Second)
	step("Doing stuff 10...", 2*time.Second)
	step("Doing stuff 11...", 2*time.Second)
	step("Doing stuff 12...", 1*time.Second)
	step("Doing stuff 13...", 5*time.Second)
	step("Doing stuff 14...", 4*time.Second)
	step("Doing stuff 15...", 8*time.Second)
	step("Doing stuff 16...", 10*time.Second)
	step("Doing stuff 17...", 1*time.Second)
	step("Done...", 2*time.Second)
	logger.done()
}

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
	subs := logger.subscribe()
	defer func() {
		log.Printf("Unsubscribing: %s", subs.id)
		subs.unsubscribe()
	}()
	w.Header().Set("X-Content-Type-Options", "nosniff")
	select {
	case startage <- struct{}{}:
	default:
	}
	for m := range subs.events {
		switch msg := m.(type) {
		case jobHistory:
			for _, i := range msg.history {
				_, err := fmt.Fprintf(w, "%s\n", i.Entry)
				if err != nil {
					return
				}
				flusher.Flush() // Trigger "chunked" encoding and send a chunk...
			}
		case jobLogEntry:
			_, err := fmt.Fprintf(w, "%s\n", msg.Entry)
			if err != nil {
				return
			}
			flusher.Flush() // Trigger "chunked" encoding and send a chunk...
		}
	}
}

func doAThing(ctx context.Context, wg *sync.WaitGroup) {
	log.Print("Starting doAThing...")
	time.Sleep(5 * time.Second)
	<-ctx.Done()
	log.Print("Finishing doAThing")
	wg.Done()
}

func waitAThing(done chan struct{}, wg *sync.WaitGroup) {
	log.Print("Waiting on doAThing")
	wg.Wait()
	log.Print("doAThing seems to have stopped")
	close(done)
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	done := make(chan struct{})
	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)

	log.Print("Starting a background thinger")
	go doAThing(ctx, &wg)
	log.Print("Starting 'wait' on a background thinger")
	go waitAThing(done, &wg)

	// Tell the thinger to stop
	cancel()
	// Wait for the "waiter" to stop
	<-done
	log.Print("All things are effectively shut down")

	go tehJob()
	log.Printf("Starting web server on port %d", port)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/stream", streamPageHandler)
	http.HandleFunc("/stream-data", streamDataHandler)

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
