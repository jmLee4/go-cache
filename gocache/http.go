package gocache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"gocache/consistenthash"
	pb "gocache/gocachepb"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_go_cache/"
	defaultReplicas = 50
)

type CacheServer struct {
	selfName         string
	basePath         string
	mu               sync.Mutex
	consistentHash   *consistenthash.ConsistentHash
	addr2CacheClient map[string]*cacheClient
}

var _ PeerPicker = (*CacheServer)(nil)

func NewCacheServer(selfName string) *CacheServer {
	return &CacheServer{
		selfName: selfName,
		basePath: defaultBasePath,
	}
}

func (p *CacheServer) Log(format string, v ...interface{}) {
	log.Printf("[GoCacheServer %s] %s", p.selfName, fmt.Sprintf(format, v...))
}

func (p *CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("CacheServer serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	// /<BasePath>/<GroupName>/<Key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: value.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(body)
}

func (p *CacheServer) Init(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.consistentHash = consistenthash.New(defaultReplicas, nil)
	p.consistentHash.InitNodes(peers...)
	p.addr2CacheClient = make(map[string]*cacheClient, len(peers))
	for _, peer := range peers {
		p.addr2CacheClient[peer] = &cacheClient{baseURL: peer + p.basePath}
	}
}

func (p *CacheServer) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.consistentHash.GetNode(key); peer != "" && peer != p.selfName {
		p.Log("Pick peer %s", peer)
		return p.addr2CacheClient[peer], true
	}
	return nil, false
}

type cacheClient struct {
	baseURL string
}

var _ PeerGetter = (*cacheClient)(nil)

func (h *cacheClient) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("unmarshal response body failed: %v", err)
	}

	return nil
}
