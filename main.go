package main

import (
	"fmt"
	"net/http"
)

var store *URLStore

const AddForm = `
<form method="POST" action="/add">
URL: <input type="text" name="url">
<input type="submit" value="Add">
</form>
`

func main() {
	store = NewURLStore("store.json")

	fmt.Println("running")

	http.HandleFunc("/", redirect)
	http.HandleFunc("/add", add)

	http.HandleFunc("/hello", hello)

	http.ListenAndServe(":8080", nil)
}

func redirect(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[1:]
	url := store.Get(key)

	if url == "" {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func add(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, AddForm)
		return
	}

	const hostname = "localhost:8080"

	key := store.Put(url)
	fmt.Fprintf(w, "http://%s/%s", hostname, key)
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
}
