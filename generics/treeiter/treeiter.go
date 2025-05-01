//go:build !solution

package treeiter

type BinTreeNode[T any] interface {
	Left() *T
	Right() *T
}

func DoInOrder[T BinTreeNode[T]](root *T, foo func(*T)) {
	if root == nil {
		return
	}
	DoInOrder((*root).Left(), foo)
	foo(root)
	DoInOrder((*root).Right(), foo)
}
