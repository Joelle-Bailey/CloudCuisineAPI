package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db := database{"shoes": 50, "socks": 5}
	mux := http.NewServeMux()
	mux.HandleFunc("/list", http.HandlerFunc(db.list))
	mux.HandleFunc("/list2", http.HandlerFunc(db.list))
	// mux.HandleFunc("/create", db.create)
	// mux.HandleFunc("/update", db.update)
	log.Fatal(http.ListenAndServe(":8000", mux))

}

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

type database map[string]dollars

func (db database) list(w http.ResponseWriter, req *http.Request) {
	for item, price := range db {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}
func (db database) list2(w http.ResponseWriter, req *http.Request) {
	for item, price := range db {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}

func (db database) price(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	price, ok := db[item]
	if !ok {

		w.WriteHeader(http.StatusNotFound) // 404
		fmt.Fprintf(w, "no such item: %q\n", item)
		return

	}
	fmt.Fprintf(w, "%s\n", price)

}


func (db database) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/list":
		for item, price := range db {
			fmt.Fprintf(w, "%s: %s\n", item, price)
		}
	case "/list2":
		for item, price := range db {
			fmt.Fprintf(w, "%s: %s\n", item, price)
		}
	case "/price":
		item := req.URL.Query().Get("item")
		price, ok := db[item]
		if !ok {
			w.WriteHeader(http.StatusNotFound) // 404
			fmt.Fprintf(w, "no such item: %q\n", item)
			return
		}
		fmt.Fprintf(w, "%s\n", price)

	default:
		w.WriteHeader(http.StatusNotFound) //404
		fmt.Fprintf(w, "no such page: %s\n", req.URL)

	}
}

// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"sync"
// )

// type dollars float32

// func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

// type database struct {
// 	data map[string]dollars
// 	mu   sync.Mutex
// }

// func newDatabase() *database {
// 	return &database{
// 		data: make(map[string]dollars),
// 	}
// }

// type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) { // adapter function
// 	f(w, r)
// }

// func (db database) list(w http.ResponseWriter, r *http.Request) {
// 	for item, price := range db.data {
// 		fmt.Fprintf(w, "%s: %s\n", item, price)
// 	}
// }

// func (db database) price(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")
// 	price, ok := db.data[item]
// 	if !ok {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "no such item: %q\n", item)
// 		return
// 	}
// 	fmt.Fprintf(w, "%s\n", price)
// }

// func (db database) create(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")
// 	price := r.URL.Query().Get("price")

// 	if _, exists := db.data[item]; exists {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "item already exists: %q\n", item)
// 		return
// 	}

// 	price_update, ok := strconv.ParseFloat(price, 64)

// 	if ok != nil {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "could not convert price: %q\n", item)
// 		return
// 	}

// 	db.mu.Lock()
// 	defer db.mu.Unlock()
// 	db.data[item] = dollars(price_update)

// 	fmt.Fprintf(w, "Created item: %s at $%.2f price\n", item, price_update)

// }

// func (db database) update(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")
// 	price := r.URL.Query().Get("price")

// 	if _, exists := db.data[item]; !exists {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "no such item: %q\n", item)
// 		return
// 	}

// 	price_update, ok := strconv.ParseFloat(price, 64)

// 	if ok != nil {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "could not convert price: %q\n", item)
// 		return
// 	}

// 	db.mu.Lock()
// 	defer db.mu.Unlock()
// 	db.data[item] = dollars(price_update)

// 	fmt.Fprintf(w, "Updated %s to %.2f\n", item, price_update)
// }

// func (db database) remove(w http.ResponseWriter, r *http.Request) {
// 	item := r.URL.Query().Get("item")

// 	if _, exists := db.data[item]; !exists {
// 		w.WriteHeader(http.StatusNotFound) //404 page
// 		fmt.Fprintf(w, "no such item: %q\n", item)
// 		return
// 	}

// 	db.mu.Lock()
// 	defer db.mu.Unlock()
// 	delete(db.data, item)

// 	fmt.Fprintf(w, "Deleted item: %s ", item)

// }

// func main() {
// 	db := newDatabase()
// 	db.data["shoes"] = 50
// 	db.data["socks"] = 5

// 	mux := http.NewServeMux()
// 	mux.Handle("/list", http.HandlerFunc(db.list))
// 	mux.Handle("/price", http.HandlerFunc(db.price))
// 	mux.Handle("/create", http.HandlerFunc(db.create))
// 	mux.Handle("/update", http.HandlerFunc(db.update))
// 	mux.Handle("/remove", http.HandlerFunc(db.remove))
// 	log.Fatal(http.ListenAndServe("localhost:8000", mux))
// }
