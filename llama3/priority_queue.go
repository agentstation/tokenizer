package llama3

import "container/heap"

// mergeNode represents a position in the token sequence that can be merged.
type mergeNode struct {
	origPos       int     // Original position in the sequence
	tokenID       int     // Token ID at this position
	mergePrio     float64 // Merge priority (lower is better)
	mergeToString string  // Result of merging with next token
	prev          *mergeNode
	next          *mergeNode
	deleted       bool // Whether this node has been deleted
	heapIndex     int  // Index in the heap (for container/heap)
}

// priorityQueue implements a min-heap of merge nodes.
type priorityQueue []*mergeNode

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	// Lower merge priority value means higher priority
	return pq[i].mergePrio < pq[j].mergePrio
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].heapIndex = i
	pq[j].heapIndex = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*mergeNode)
	node.heapIndex = n
	*pq = append(*pq, node)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // avoid memory leak
	node.heapIndex = -1
	*pq = old[0 : n-1]
	return node
}

// newPriorityQueue creates a new priority queue for merge operations.
func newPriorityQueue() *priorityQueue {
	pq := &priorityQueue{}
	heap.Init(pq)
	return pq
}
