package frmtr

import (
	"sort"
	"strings"
)

type Files []string

func (f Files) String() string {
	root := &node{
		name:     ".",
		children: make(map[string]*node),
	}

	for _, path := range f {
		parts := strings.Split(path, "/") // TODO:
		root.addPath(parts)
	}

	return root.string("", true, true)
}

type node struct {
	name     string
	children map[string]*node
}

func (n *node) addPath(parts []string) {
	if len(parts) == 0 {
		return
	}

	part := parts[0]
	child, exists := n.children[part]
	if !exists {
		child = &node{
			name:     part,
			children: make(map[string]*node),
		}
		n.children[part] = child
	}

	child.addPath(parts[1:])
}

func (n *node) string(prefix string, isRoot, isLast bool) string {
	var sb strings.Builder

	if isRoot {
		sb.WriteString(n.name + "\n") // Root has no prefix
	} else {
		if isLast {
			sb.WriteString(prefix + "└── " + n.name + "\n")
		} else {
			sb.WriteString(prefix + "├── " + n.name + "\n")
		}
	}

	keys := make([]string, 0, len(n.children))
	for k := range n.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		child := n.children[k]

		newPrefix := prefix
		if !isRoot {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}

		isLastChild := i == len(keys)-1

		sb.WriteString(child.string(newPrefix, false, isLastChild))
	}

	return sb.String()
}
