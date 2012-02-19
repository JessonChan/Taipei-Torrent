package dht

import (
	"log"
)

type nTree struct {
	left, right, parent *nTree
	value               *DhtRemoteNode
}

const (
	idLen = 20
	// Each query returns up to this number of nodes.
	kNodes = 8
)

func (n *nTree) insert(newNode *DhtRemoteNode) {
	id := newNode.id
	if len(id) != idLen {
		return
	}
	var bit uint
	var chr byte
	next := n
	for i := 0; i < len(id); i++ {
		chr = id[i]
		for bit = 0; bit < 8; bit++ {
			if chr>>bit&1 == 1 {
				if next.right == nil {
					next.right = &nTree{parent: next}
				}
				next = next.right
			} else {
				if next.left == nil {
					next.left = &nTree{parent: next}
				}
				next = next.left
			}
		}
	}
	if next.value != nil && next.value.id == id {
		// There's already a node with this id. Keep.
		return
	}
	next.value = newNode
}

func (n *nTree) lookupClosest(id string) []*DhtRemoteNode {
	// Find value, or neighbors up to kNodes.
	next := n
	var bit uint
	var chr byte
	for i := 0; i < len(id); i++ {
		chr = id[i]
		for bit = 0; bit < 8; bit++ {
			if chr>>bit&1 == 1 {
				if next.right == nil {
					// Reached bottom of the match tree. Start going backwards.
					return next.left.reverse()
				}
				next = next.right
			} else {
				if next.left == nil {
					return next.right.reverse()
				}
				next = next.left
			}
		}
	}
	// Found exact match.
	// XXX: This is not correct. We actually want to return kNodes for this too.
	return []*DhtRemoteNode{next.value}
}

func (n *nTree) reverse() []*DhtRemoteNode {
	ret := make([]*DhtRemoteNode, 0, kNodes)
	var back *nTree
	node := n

	for {
		if len(ret) >= kNodes {
			return ret
		}
		// Don't go down the same branch we came from.
		if node.right != nil && node.right != back {
			ret = node.right.everything(ret)
		} else if node.left != nil && node.left != back {
			ret = node.left.everything(ret)
		}
		if node.parent == nil {
			// Reached top of the tree.
			break
		}
		back = node
		node = node.parent
	}
	return ret // Partial results :-(.
}

// evertyhing traverses the whole tree and collects up to
// kNodes values, without any ordering guarantees.
func (n *nTree) everything(ret []*DhtRemoteNode) []*DhtRemoteNode {
	if n.value != nil {
		if n.value.reachable {
			return append(ret, n.value)
		}
		log.Printf("Node %x not reachable. Ignoring.", n.value.id)
		return ret
	}
	if len(ret) >= kNodes {
		goto RET
	}
	if n.right != nil {
		ret = n.right.everything(ret)
	}
	if n.left != nil {
		ret = n.left.everything(ret)
	}
RET:
	if len(ret) > kNodes {
		ret = ret[0:kNodes]
	}
	return ret
}
