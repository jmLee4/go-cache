package main

import (
	"flag"
	"fmt"
	"gocache"
	"log"
	"net/http"
)

const (
	APIServerRoute   = "/api"
	APIServerKey     = "filename"
	defaultGroupName = "static_files"
)

var staticFiles = []string{
	"apple", "flower", "horse", "house", "panda", "tiger", "tree", "man", "light",
}

var slowDB map[string]*gocache.StaticFile

func initSlowDB() {
	slowDB = make(map[string]*gocache.StaticFile)
	for _, filename := range staticFiles {
		slowDB[filename] = gocache.NewStaticFile(filename)
	}
}

func createGroup(addrInfo string) *gocache.Group {
	return gocache.NewGroup(addrInfo, defaultGroupName, 3<<20, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] Search key", key)
			if v, ok := slowDB[key]; ok {
				return v.GetContent(), nil
			}
			log.Printf("[SlowDB] %s not exist", key)
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(cacheServerAddr string, allCacheServerAddr []string, group *gocache.Group) {
	cacheServer := gocache.NewCacheServer(cacheServerAddr)
	cacheServer.Init(allCacheServerAddr...)
	group.RegisterPeers(cacheServer)
	log.Println("GoCacheServer is running at", cacheServerAddr)
	log.Fatal(http.ListenAndServe(cacheServerAddr[7:], cacheServer))
}

func startAPIServer(apiAddr string, group *gocache.Group) {
	http.Handle(APIServerRoute, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get(APIServerKey)
			log.Println()
			log.Println("[APIServer] Get query for key:", key)
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.jpg\"", key))
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write(view.ByteSlice())
		}))
	log.Println("APIServer is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var cacheServerPort int
	var startAPIServerFlag bool
	flag.IntVar(&cacheServerPort, "cacheServerPort", 8001, "GoCache server port")
	flag.BoolVar(&startAPIServerFlag, "startAPIServerFlag", false, "Start a APIServer?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	port2Addr := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var allCacheServerAddr []string
	for _, addr := range port2Addr {
		allCacheServerAddr = append(allCacheServerAddr, addr)
	}

	initSlowDB()

	group := createGroup(port2Addr[cacheServerPort])
	if startAPIServerFlag {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(port2Addr[cacheServerPort], allCacheServerAddr, group)
}
