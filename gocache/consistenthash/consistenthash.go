package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFn func(data []byte) uint32

type ConsistentHash struct {
	hashFn     HashFn
	replicas   int
	values     []int
	value2Node map[int]string
}

func New(replicas int, fn HashFn) *ConsistentHash {
	ct := &ConsistentHash{
		replicas:   replicas,
		hashFn:     fn,
		value2Node: make(map[int]string),
	}
	if ct.hashFn == nil {
		ct.hashFn = crc32.ChecksumIEEE
	}
	return ct
}

func (m *ConsistentHash) InitNodes(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < m.replicas; i++ {
			value := int(m.hashFn([]byte(strconv.Itoa(i) + node)))
			m.values = append(m.values, value)
			m.value2Node[value] = node
		}
	}
	sort.Ints(m.values)
}

func (m *ConsistentHash) GetNode(key string) string {
	if len(m.values) == 0 {
		return ""
	}

	value := int(m.hashFn([]byte(key)))
	idx := sort.Search(len(m.values), func(i int) bool {
		return m.values[i] >= value
	})
	return m.value2Node[m.values[idx%len(m.values)]]
}
