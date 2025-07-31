package llama3

import (
	"container/heap"
	"strings"
)

// performBPE executes the Byte Pair Encoding algorithm on a pre-token.
func (t *Tokenizer) performBPE(pretoken string) []int {
	// Check cache first
	t.cacheMutex.RLock()
	if cached, ok := t.cache[pretoken]; ok {
		t.cacheMutex.RUnlock()
		return cached
	}
	t.cacheMutex.RUnlock()
	
	// Check if the entire pretoken exists in vocabulary (ignore_merges behavior)
	if tokenID, ok := t.vocabByString[pretoken]; ok {
		result := []int{tokenID}
		t.cacheResult(pretoken, result)
		return result
	}
	
	// Convert pretoken to initial token IDs (one per character)
	tokenIDs := make([]int, 0, len(pretoken))
	for _, r := range pretoken {
		char := string(r)
		if id, ok := t.vocabByString[char]; ok {
			tokenIDs = append(tokenIDs, id)
		} else {
			// This can happen with some special Unicode characters
			// Skip the character rather than panic
			continue
		}
	}
	
	// If no valid tokens were found, return empty
	if len(tokenIDs) == 0 {
		t.cacheResult(pretoken, tokenIDs)
		return tokenIDs
	}
	
	// If only one token, no merging needed
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
		if leftOfMerge.deleted || leftOfMerge.next == nil || leftOfMerge.next.deleted {
			continue
		}
		
		// Mark nodes as deleted
		leftOfMerge.deleted = true
		leftOfMerge.next.deleted = true
		
		// Handle the previous node
		if leftOfMerge.prev != nil {
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
		}
		
		// Create merged node
		mergedTokenID, ok := t.vocabByString[leftOfMerge.mergeToString]
		if !ok {
			// This shouldn't happen with valid merge data
			continue
		}
		
		resultOfMerge := &mergeNode{
			origPos: leftOfMerge.origPos,
			tokenID: mergedTokenID,
			prev:    leftOfMerge.prev,
			next:    leftOfMerge.next.next,
		}
		
		// Update links and add new merge possibilities
		if resultOfMerge.prev != nil {
			resultOfMerge.prev.next = resultOfMerge
			t.addToMergeQueue(resultOfMerge.prev, pq, len(pretoken))
		} else {
			firstNode = resultOfMerge
		}
		
		if resultOfMerge.next != nil {
			resultOfMerge.next.prev = resultOfMerge
			t.addToMergeQueue(resultOfMerge, pq, len(pretoken))
		}
	}
	
	// Collect final token IDs
	result := make([]int, 0)
	for node := firstNode; node != nil; node = node.next {
		result = append(result, node.tokenID)
	}
	
	t.cacheResult(pretoken, result)
	return result
}

// buildMergeList creates the initial linked list and populates the priority queue.
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

// addToMergeQueue adds a potential merge to the priority queue.
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

// cacheResult stores a BPE result in the cache.
func (t *Tokenizer) cacheResult(pretoken string, result []int) {
	t.cacheMutex.Lock()
	t.cache[pretoken] = result
	t.cacheMutex.Unlock()
}