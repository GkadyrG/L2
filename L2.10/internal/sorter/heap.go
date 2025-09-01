package sorter

type QueueElement struct {
	content     string
	sourceIndex int
}

type MinimalHeap []QueueElement

func (h MinimalHeap) Len() int           { return len(h) }
func (h MinimalHeap) Less(i, j int) bool { return h[i].content < h[j].content }
func (h MinimalHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinimalHeap) Push(element any) { *h = append(*h, element.(QueueElement)) }
func (h *MinimalHeap) Pop() any {
	current := *h
	length := len(current)
	element := current[length-1]
	*h = current[:length-1]
	return element
}
