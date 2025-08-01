package bpe

import (
	"container/heap"
	"strings"
)

// Processor handles the BPE algorithm implementation.
type Processor struct {
	Tokens      []string       // Token ID to text mapping
	TokenLookup map[string]int // Text to token ID mapping
	MergeRules  map[string]int // BPE merge rules with priorities
	Cache       Cache          // Cache for BPE results
}

// PerformBPE executes the Byte Pair Encoding algorithm on a pre-token.
func (p *Processor) PerformBPE(pretoken string) []int {
	// Try to get from cache
	if p.Cache != nil {
		if cached, ok := p.Cache.Get(pretoken); ok {
			return cached
		}
	}

	// Check for direct vocabulary match
	if tokenID, ok := p.TokenLookup[pretoken]; ok {
		result := []int{tokenID}
		if p.Cache != nil {
			p.Cache.Put(pretoken, result)
		}
		return result
	}

	// Convert to initial tokens
	tokenIDs := p.pretokenToIDs(pretoken)
	if len(tokenIDs) <= 1 {
		if p.Cache != nil {
			p.Cache.Put(pretoken, tokenIDs)
		}
		return tokenIDs
	}

	// Build linked list and priority queue
	pq := NewPriorityQueue()
	firstNode := p.buildMergeList(tokenIDs, pq, len(pretoken))

	// Perform merges
	for pq.Len() > 0 {
		leftOfMerge := heap.Pop(pq).(*MergeNode)

		// Skip if this merge is no longer valid
		if !p.isValidMerge(leftOfMerge) {
			continue
		}

		// Perform the merge
		firstNode = p.performMerge(leftOfMerge, firstNode, pq, len(pretoken))
	}

	// Collect final token IDs
	result := make([]int, 0)
	for node := firstNode; node != nil; node = node.Next {
		result = append(result, node.TokenID)
	}

	if p.Cache != nil {
		p.Cache.Put(pretoken, result)
	}
	return result
}

// pretokenToIDs converts a pretoken string to initial token IDs (one per character).
func (p *Processor) pretokenToIDs(pretoken string) []int {
	tokenIDs := make([]int, 0, len(pretoken))
	for _, r := range pretoken {
		char := string(r)
		if id, ok := p.TokenLookup[char]; ok {
			tokenIDs = append(tokenIDs, id)
		}
		// Skip characters not in vocabulary
	}
	return tokenIDs
}

// buildMergeList creates the initial linked list of tokens and populates
// the priority queue with possible merges.
func (p *Processor) buildMergeList(tokenIDs []int, pq *PriorityQueue, pretokenLen int) *MergeNode {
	if len(tokenIDs) == 0 {
		return nil
	}

	firstNode := &MergeNode{
		OrigPos: 0,
		TokenID: tokenIDs[0],
	}

	prevNode := firstNode
	for i := 1; i < len(tokenIDs); i++ {
		currNode := &MergeNode{
			OrigPos: i,
			TokenID: tokenIDs[i],
			Prev:    prevNode,
		}
		prevNode.Next = currNode
		p.addToMergeQueue(prevNode, pq, pretokenLen)
		prevNode = currNode
	}

	return firstNode
}

// addToMergeQueue evaluates a pair of adjacent tokens for merging.
func (p *Processor) addToMergeQueue(leftNode *MergeNode, pq *PriorityQueue, pretokenLen int) {
	if leftNode.Next == nil {
		return
	}

	mergeIdentifier := p.getMergeIdentifier(leftNode.TokenID, leftNode.Next.TokenID)
	mergePrio, ok := p.MergeRules[mergeIdentifier]
	if !ok {
		return // This merge is not possible
	}

	// Add position bias to ensure left-to-right processing of equal priority merges
	leftNode.MergePrio = float64(mergePrio) + float64(leftNode.OrigPos)/float64(pretokenLen)
	leftNode.MergeToString = strings.ReplaceAll(mergeIdentifier, " ", "")

	heap.Push(pq, leftNode)
}

// getMergeIdentifier creates a merge identifier from two token IDs.
func (p *Processor) getMergeIdentifier(firstTokenID, secondTokenID int) string {
	if firstTokenID >= len(p.Tokens) || secondTokenID >= len(p.Tokens) {
		return ""
	}
	return p.Tokens[firstTokenID] + " " + p.Tokens[secondTokenID]
}

// isValidMerge checks if a merge node is still valid for merging.
func (p *Processor) isValidMerge(node *MergeNode) bool {
	return node != nil && !node.Deleted && node.Next != nil && !node.Next.Deleted
}

// performMerge executes a single merge operation and updates the linked list.
func (p *Processor) performMerge(leftOfMerge *MergeNode, firstNode *MergeNode, pq *PriorityQueue, pretokenLen int) *MergeNode {
	// Mark nodes as deleted
	leftOfMerge.Deleted = true
	leftOfMerge.Next.Deleted = true

	// Handle the previous node
	if leftOfMerge.Prev != nil {
		firstNode = p.updatePreviousNode(leftOfMerge, firstNode)
	}

	// Create merged node
	mergedTokenID, ok := p.TokenLookup[leftOfMerge.MergeToString]
	if !ok {
		// This shouldn't happen with valid merge data
		return firstNode
	}

	resultOfMerge := &MergeNode{
		OrigPos: leftOfMerge.OrigPos,
		TokenID: mergedTokenID,
		Prev:    leftOfMerge.Prev,
		Next:    leftOfMerge.Next.Next,
	}

	// Update links and add new merge possibilities
	firstNode = p.updateMergeLinks(resultOfMerge, firstNode, pq, pretokenLen)

	return firstNode
}

// updatePreviousNode handles updating the previous node during a merge.
func (p *Processor) updatePreviousNode(leftOfMerge *MergeNode, firstNode *MergeNode) *MergeNode {
	oldPrev := leftOfMerge.Prev
	oldPrev.Deleted = true

	// Create new previous node
	newPrev := &MergeNode{
		OrigPos: oldPrev.OrigPos,
		TokenID: oldPrev.TokenID,
		Prev:    oldPrev.Prev,
		Next:    oldPrev.Next,
	}
	leftOfMerge.Prev = newPrev

	if newPrev.Prev != nil {
		newPrev.Prev.Next = newPrev
	} else {
		firstNode = newPrev
	}

	return firstNode
}

// updateMergeLinks updates the linked list links after a merge.
func (p *Processor) updateMergeLinks(resultOfMerge *MergeNode, firstNode *MergeNode, pq *PriorityQueue, pretokenLen int) *MergeNode {
	if resultOfMerge.Prev != nil {
		resultOfMerge.Prev.Next = resultOfMerge
		p.addToMergeQueue(resultOfMerge.Prev, pq, pretokenLen)
	} else {
		firstNode = resultOfMerge
	}

	if resultOfMerge.Next != nil {
		resultOfMerge.Next.Prev = resultOfMerge
		p.addToMergeQueue(resultOfMerge, pq, pretokenLen)
	}

	return firstNode
}
