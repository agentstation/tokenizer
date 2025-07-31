package llama3

import (
	"container/heap"
	"strings"
)

// performBPE executes the Byte Pair Encoding algorithm on a pre-token.
// It iteratively merges the most frequent pairs of adjacent tokens according
// to the learned merge rules. Results are cached for efficiency.
func (t *Tokenizer) performBPE(pretoken string) []int {
	// Try to get from cache
	if cached := t.getCached(pretoken); cached != nil {
		return cached
	}

	// Check for direct vocabulary match
	if result := t.tryDirectMatch(pretoken); result != nil {
		return result
	}

	// Convert to initial tokens
	tokenIDs := t.pretokenToIDs(pretoken)
	if len(tokenIDs) <= 1 {
		t.cacheResult(pretoken, tokenIDs)
		return tokenIDs
	}

	// Build linked list and priority queue
	pq := newPriorityQueue()
	firstNode := t.buildMergeList(tokenIDs, pq, len(pretoken))

	// Perform merges
	for pq.Len() > 0 {
		leftOfMerge := heap.Pop(pq).(*mergeNode)

		// Skip if this merge is no longer valid
		if !t.isValidMerge(leftOfMerge) {
			continue
		}

		// Perform the merge
		firstNode = t.performMerge(leftOfMerge, firstNode, pq, len(pretoken))
	}

	// Collect final token IDs
	result := make([]int, 0)
	for node := firstNode; node != nil; node = node.next {
		result = append(result, node.tokenID)
	}

	t.cacheResult(pretoken, result)
	return result
}

// buildMergeList creates the initial linked list of tokens and populates
// the priority queue with possible merges. Each token becomes a node in the
// linked list, and adjacent pairs are evaluated for merging.
func (t *Tokenizer) buildMergeList(tokenIDs []int, pq *priorityQueue, pretokenLen int) *mergeNode {
	if len(tokenIDs) == 0 {
		return nil
	}

	firstNode := &mergeNode{
		origPos: 0,
		tokenID: tokenIDs[0],
	}

	prevNode := firstNode
	for i := 1; i < len(tokenIDs); i++ {
		currNode := &mergeNode{
			origPos: i,
			tokenID: tokenIDs[i],
			prev:    prevNode,
		}
		prevNode.next = currNode
		t.addToMergeQueue(prevNode, pq, pretokenLen)
		prevNode = currNode
	}

	return firstNode
}

// addToMergeQueue evaluates a pair of adjacent tokens for merging and adds
// them to the priority queue if a valid merge exists. The merge priority is
// adjusted by position to ensure deterministic left-to-right processing.
func (t *Tokenizer) addToMergeQueue(leftNode *mergeNode, pq *priorityQueue, pretokenLen int) {
	if leftNode.next == nil {
		return
	}

	mergeIdentifier := t.getMergeIdentifier(leftNode.tokenID, leftNode.next.tokenID)
	mergePrio, ok := t.merges[mergeIdentifier]
	if !ok {
		return // This merge is not possible
	}

	// Add position bias to ensure left-to-right processing of equal priority merges
	leftNode.mergePrio = float64(mergePrio) + float64(leftNode.origPos)/float64(pretokenLen)
	leftNode.mergeToString = strings.ReplaceAll(mergeIdentifier, " ", "")

	heap.Push(pq, leftNode)
}

// getCached retrieves a cached BPE result if available.
func (t *Tokenizer) getCached(pretoken string) []int {
	if t.cache != nil {
		if cached, ok := t.cache.get(pretoken); ok {
			return cached
		}
	}
	return nil
}

// tryDirectMatch checks if the entire pretoken exists in the vocabulary.
func (t *Tokenizer) tryDirectMatch(pretoken string) []int {
	if tokenID, ok := t.vocabByString[pretoken]; ok {
		result := []int{tokenID}
		t.cacheResult(pretoken, result)
		return result
	}
	return nil
}

// pretokenToIDs converts a pretoken string to initial token IDs (one per character).
func (t *Tokenizer) pretokenToIDs(pretoken string) []int {
	tokenIDs := make([]int, 0, len(pretoken))
	for _, r := range pretoken {
		char := string(r)
		if id, ok := t.vocabByString[char]; ok {
			tokenIDs = append(tokenIDs, id)
		}
		// Skip characters not in vocabulary
	}
	return tokenIDs
}

// cacheResult stores the BPE encoding result in the cache for future lookups.
// This significantly improves performance for repeated tokens.
func (t *Tokenizer) cacheResult(pretoken string, result []int) {
	if t.cache != nil {
		t.cache.put(pretoken, result)
	}
}

// isValidMerge checks if a merge node is still valid for merging.
func (t *Tokenizer) isValidMerge(node *mergeNode) bool {
	return node != nil && !node.deleted && node.next != nil && !node.next.deleted
}

// performMerge executes a single merge operation and updates the linked list.
func (t *Tokenizer) performMerge(leftOfMerge *mergeNode, firstNode *mergeNode, pq *priorityQueue, pretokenLen int) *mergeNode {
	// Mark nodes as deleted
	leftOfMerge.deleted = true
	leftOfMerge.next.deleted = true

	// Handle the previous node
	if leftOfMerge.prev != nil {
		firstNode = t.updatePreviousNode(leftOfMerge, firstNode)
	}

	// Create merged node
	mergedTokenID, ok := t.vocabByString[leftOfMerge.mergeToString]
	if !ok {
		// This shouldn't happen with valid merge data
		return firstNode
	}

	resultOfMerge := &mergeNode{
		origPos: leftOfMerge.origPos,
		tokenID: mergedTokenID,
		prev:    leftOfMerge.prev,
		next:    leftOfMerge.next.next,
	}

	// Update links and add new merge possibilities
	firstNode = t.updateMergeLinks(resultOfMerge, firstNode, pq, pretokenLen)
	
	return firstNode
}

// updatePreviousNode handles updating the previous node during a merge.
func (t *Tokenizer) updatePreviousNode(leftOfMerge *mergeNode, firstNode *mergeNode) *mergeNode {
	oldPrev := leftOfMerge.prev
	oldPrev.deleted = true

	// Create new previous node
	newPrev := &mergeNode{
		origPos: oldPrev.origPos,
		tokenID: oldPrev.tokenID,
		prev:    oldPrev.prev,
		next:    oldPrev.next,
	}
	leftOfMerge.prev = newPrev

	if newPrev.prev != nil {
		newPrev.prev.next = newPrev
	} else {
		firstNode = newPrev
	}
	
	return firstNode
}

// updateMergeLinks updates the linked list links after a merge.
func (t *Tokenizer) updateMergeLinks(resultOfMerge *mergeNode, firstNode *mergeNode, pq *priorityQueue, pretokenLen int) *mergeNode {
	if resultOfMerge.prev != nil {
		resultOfMerge.prev.next = resultOfMerge
		t.addToMergeQueue(resultOfMerge.prev, pq, pretokenLen)
	} else {
		firstNode = resultOfMerge
	}

	if resultOfMerge.next != nil {
		resultOfMerge.next.prev = resultOfMerge
		t.addToMergeQueue(resultOfMerge, pq, pretokenLen)
	}
	
	return firstNode
}
