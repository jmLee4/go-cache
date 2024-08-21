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
	APIServerKey     = "key"
	DefaultGroupName = "scores"
)

var slowDB = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gocache.Group {
	return gocache.NewGroup(DefaultGroupName, 5<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] Search key", key)
			if v, ok := slowDB[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(cacheServerAddr string, allCacheServerAddr []string, group *gocache.Group) {
	cacheServer := gocache.NewHTTPPool(cacheServerAddr)
	cacheServer.Set(allCacheServerAddr...)
	group.RegisterPeers(cacheServer)
	log.Println("GoCacheServer is running at", cacheServerAddr)
	log.Fatal(http.ListenAndServe(cacheServerAddr[7:], cacheServer))
}

func startAPIServer(apiAddr string, group *gocache.Group) {
	http.Handle(APIServerRoute, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get(APIServerKey)
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// TODO: Change file name
			w.Header().Set("Content-Disposition", "attachment; filename=\"temp.txt\"")
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

	group := createGroup()
	if startAPIServerFlag {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(port2Addr[cacheServerPort], allCacheServerAddr, group)
}
