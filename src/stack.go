package azb

type Stack struct {
	top  *stackNode
	size int
}

type stackNode struct {
	value string
	next  *stackNode
}

// Return the stack's length
func (es *Stack) Len() int {
	return es.size
}

// Push a new element onto the stack
func (es *Stack) Push(value ...string) {
	for _, v := range value {
		es.top = &stackNode{v, es.top}
		es.size++
	}
}

// Remove the top element from the stack and return its value
// If the stack is empty, return NONE
func (es *Stack) Pop() (value string, ok bool) {
	if es.size > 0 {
		value, es.top = es.top.value, es.top.next
		es.size--
		return value, true
	}

	return "", false
}

// Looks at the top item on the stack, and returns its value
// Does not pop that item
func (es *Stack) Peek() (value string, ok bool) {
	if es.size > 0 {
		return es.top.value, true
	}

	return "", false
}

// String representation of the stack
func (es *Stack) String() string {
	str := ""
	item := es.top
	for {
		if item == nil {
			return str
		}

		if str != "" {
			str = " " + str
		}
		str = "{" + string(item.value) + "}" + str
		item = item.next
	}
}

func (es *Stack) Array() []string {
	if es.Len() == 0 {
		return []string{}
	}

	arr := make([]string, es.Len())

	i := 0
	item := es.top
	for {
		if item == nil {
			return arr
		}

		arr[i] = item.value

		item = item.next

		i++
	}
}

func (es *Stack) Reverse() []string {
	if es.Len() == 0 {
		return []string{}
	}

	arr := make([]string, es.Len())

	i := es.Len() - 1
	item := es.top
	for {
		if item == nil {
			return arr
		}

		arr[i] = item.value

		item = item.next

		i--
	}
}
