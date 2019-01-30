package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	forwardedHeader  = "X-Forwarded-Proto"
	httpsPort        = "443"
	httpSecureSchema = "https"
	nonSecureSchema  = "ws"
	secureSchema     = "wss"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.ServeFile(w, r, "./home.html")
}

func serveMain(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		return
	}

	schema := nonSecureSchema
	if r.Header.Get(forwardedHeader) == httpSecureSchema {
		schema = secureSchema
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Unable to decode request body")
		return
	}

	vars := mux.Vars(r)

	// Check for JSON in the request body if not empty
	var postPayload string
	if len(body) > 0 {
		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Unable to parse valid JSON")
			return
		}

		postPayload = string(body)
	}

	params := r.URL.Query()
	if len(params) == 0 && len(postPayload) == 0 {
		homeTemplate.Execute(w, schema + "://"+r.Host+"/server/"+vars["path"])
		return
	}

	getPayload := toJSON(params)
	message := fmt.Sprintf(`{"method":"%s","headers":%s,"get":%s,"post":%s}`, r.Method, toJSON(r.Header), getPayload, postPayload)
	broadcastMessage(r, message, "/server/"+vars["path"])
}

func serveNew(w http.ResponseWriter, r *http.Request) {
	randURL := randomString(8, alphabeticalType)
	http.Redirect(w, r, "/"+randURL, http.StatusTemporaryRedirect)
}

func broadcastMessage(r *http.Request, message, path string) {
	u := url.URL{Scheme: nonSecureSchema, Host: os.Getenv("HOST") + ":" + os.Getenv("PORT"), Path: path}
	if r.Header.Get(forwardedHeader) == httpSecureSchema {
		u = url.URL{Scheme: secureSchema, Host: os.Getenv("SECURE_HOST") + ":" + httpsPort, Path: path}
	}
	log.Println("connecting to ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("error creating a new web socket client connection: ", err)
		return
	}
	defer c.Close()

	err = c.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Println("write error: ", err)
		return
	}
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()

	r := mux.NewRouter()
	r.HandleFunc("/", serveHome)
	r.HandleFunc("/new", serveNew)
	r.HandleFunc("/{path}", serveMain)
	r.HandleFunc("/server/{path}", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./"))))

	err := http.ListenAndServe(":"+os.Getenv("PORT"), r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	<link rel="stylesheet" href="//stackpath.bootstrapcdn.com/bootstrap/4.1.3/css/bootstrap.min.css" integrity="sha384-MCw98/SFnGE8fJT3GXwEOngsV7Zt27NXFoaoApmYm81iuXoPkFOJwJ8ERdknLPMO" crossorigin="anonymous">
	<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/font-awesome/4.2.0/css/font-awesome.min.css">
	<script src="//cdnjs.cloudflare.com/ajax/libs/moment.js/2.22.1/moment.min.js"></script>
    <link rel="stylesheet" href="//cdn.jsdelivr.net/gh/highlightjs/cdn-release@9.13.1/build/styles/atom-one-dark.min.css">
    <script src="//cdn.jsdelivr.net/gh/highlightjs/cdn-release@9.13.1/build/highlight.min.js"></script>

	<title>Webhookr.go</title>

	<script type="text/javascript">
		hljs.initHighlightingOnLoad();
		hljs.configure({useBR: true});

		var audio = new Audio("/static/alert.ogg")

		function removeEntry(e) {
			var childNode = e.parentNode.parentNode
			childNode.parentNode.removeChild(childNode)
		}

		function toggleHeaders(e) {
			var curState = e.parentNode.parentNode.getElementsByClassName("headers")[0].style.display;
			console.log(curState)
			if (curState == "none") {
				e.parentNode.parentNode.getElementsByClassName("headers")[0].style.display = "block"
			} else {
				e.parentNode.parentNode.getElementsByClassName("headers")[0].style.display = "none"
			}
			console.log(curState)
		}

		window.onload = function () {
			var conn;
			var msg = document.getElementById("msg");
			var log = document.getElementById("log");

			function appendLog(item) {
				var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
				log.appendChild(item);
				if (doScroll) {
					log.scrollTop = log.scrollHeight - log.clientHeight;
				}
			}

			if (window["WebSocket"]) {
				conn = new WebSocket("{{.}}");
				conn.onclose = function (evt) {
					var item = document.createElement("div");
					item.innerHTML = "<b>Connection closed.</b>";
					appendLog(item);
				};
				conn.onmessage = function (evt) {
					audio.play()
					var json = JSON.parse(evt.data)
					var item = document.createElement("div");
					item.className = "card mb-3"
					item.innerHTML = getHtmlBlock(json.method, json.headers, json.get, json.post)
					appendLog(item);
				};
			} else {
				var item = document.createElement("div");
				item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
				appendLog(item);
			}
		};

		function getHtmlBlock(method, headers, getPayload, postPayload) {
			var time = moment().format()
			var html = {
				headTemplate: [
						'<div class="card-header">',
							'<button type="button" class="btn btn-success float-left">' + method + '</button><span>&nbsp;</span>',
							'<button type="button" class="btn btn-outline-primary float-cent">' + time + '</button>',
							'<button type="button" class="btn btn-danger float-right glyphicon glyphicon-remove-circle" onclick="removeEntry(this)"><i class="fa fa-fw fa-trash"></i></button>',
						'</div>',
						'<div class="card-body">',
							'<div class="form-check">',
								'<input type="checkbox" class="form-check-input" id="showHeaders" onclick="toggleHeaders(this)">',
								'<label class="form-check-label" for="showHeaders">Show headers</label>',
							'</div>',
							'<div class="alert alert-dark headers" role="alert" id="headers" style="display: none">',
								'<pre>',
									'<code class="json">' + JSON.stringify(headers, undefined, 2) + '</code>',
								'</pre>',
							'</div>',
				].join(""),
				getTemplate: [
							'<div class="alert alert-dark" role="alert" id="getPayload">',
								'<pre>',
									'<code class="json">' + JSON.stringify(getPayload, undefined, 2) + '</code>',
								'</pre>',
							'</div>',
				].join(""),
				postTemplate: [
							'<div class="alert alert-dark" role="alert" id="postPayload">',
								'<pre>',
									'<code class="json">' + JSON.stringify(postPayload, undefined, 2) + '</code>',
								'</pre>',
							'</div>',
				].join(""),
				footTemplate: [
						'</div>',
				].join("")
			}

			var output = html.headTemplate

			if (Object.keys(getPayload).length > 0) {
				output += html.getTemplate;
			}

			if (Object.keys(postPayload).length > 0) {
				output += html.postTemplate;
			}

			output += html.footTemplate

			return output
		}
		</script>
	</head>
<body>

	<div class="container">
		<h2>Webhookr.go</h2>
		<a href="/new">New Webhookr</a>

		<div id="log">
		</div>
	</div>
</body>
</html>
`))
