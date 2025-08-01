package bpe

import "container/heap"

// MergeNode represents a position in the token sequence that can be merged.
type MergeNode struct {
	OrigPos       int     // Original position in the sequence
	TokenID       int     // Token ID at this position
	MergePrio     float64 // Merge priority (lower is better)
	MergeToString string  // Result of merging with next token
	Prev          *MergeNode
	Next          *MergeNode
	Deleted       bool // Whether this node has been deleted
	HeapIndex     int  // Index in the heap (for container/heap)
}

// PriorityQueue implements a min-heap of merge nodes.
type PriorityQueue []*MergeNode

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// Lower merge priority value means higher priority
	return pq[i].MergePrio < pq[j].MergePrio
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].HeapIndex = i
	pq[j].HeapIndex = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	node := x.(*MergeNode)
	node.HeapIndex = n
	*pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil // avoid memory leak
	node.HeapIndex = -1
	*pq = old[0 : n-1]
	return node
}

// NewPriorityQueue creates a new priority queue for merge operations.
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{}
	heap.Init(pq)
	return pq
}
