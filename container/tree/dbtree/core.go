package dbtree

import (
	"github.com/leoheung/go-patterns/container/list"
)


func contains(seen *list.List[string], item string) bool{
	return seen.Includes(item, func(a, b string) bool {return a == b})
}


func BuildDependencyOrder(tables []string, tablesDependencies map[string][]string) []string {
	if len(tables) == 0 {
		return nil
	}

	seen := list.New[string]()
	stack := list.From(tables)
	
	for temp, ok := stack.Peek(); ok ; temp,ok = stack.Peek(){
		if contains(seen, temp) {
			stack.Pop()
			continue
		}

		deps, ok := tablesDependencies[temp]
		if (!ok || len(deps) == 0) {
			if !contains(seen, temp) {
				seen.Append(temp)
				stack.Pop()
			}
		} else {
			done := true
			for _, dep := range deps {
				if !contains(seen, dep) {
					stack.Push(dep)
					done = false
				}
			}

			if done {
				stack.Pop()
				seen.Append(temp)
			}
		}
	}

	return seen.ToSlice()
}
