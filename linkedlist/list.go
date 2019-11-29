package linkedlist

import (
	"sync/atomic"
	"unsafe"
)



// Node is a Linked List Node
type Node struct {
	next    *Node

	//
	version uint64

	deleted *bool // Fucking atomic replace!


	object unsafe.Pointer
}


// LinkedList is a linked list?
type LinkedList struct {
	head *Node
}

// Insert Inserts a node into the list
func (ll *LinkedList) Insert(version uint64, data unsafe.Pointer) *Node {

	// 取出链表头 head
	currentHead := ll.head

	// 是否被删除
	f := false

	// 如果当前 ll.head 为空 或者 新数据 data 的版本号更新，就将新数据 data 插入到队列头部。
	if currentHead == nil || version > currentHead.version {

		// 构造新节点
		newNode := &Node{
			version: version, 		// 版本号
			deleted: &f,			// 是否被删除
			next:    currentHead, 	// 插入到头 head 之前
			object: data, 			// 数据
		}

		// 原子的将 ll.head 由 currentHead 更新为 newNode， 如果失败就递归调用 ll.Insert() 函数，相当于不断的重试 (while)
		if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&ll.head)), unsafe.Pointer(currentHead), unsafe.Pointer(newNode)) {
			return ll.Insert(version, data)
		}

		// 返回新插入的节点指针
		return newNode
	}

	// 如果当前 ll.head 不为空 或者 新数据 data 的版本号不比 head 新，需要将新数据 data 插入到队列中合适位置。

	// 从 ll.head 开始遍历链表
	cursor := ll.head

	for {

		// 如果遍历到链表尾，或者发现版本号适当的插入位置，就执行插入，插入位置是 cursor 之后，cursor->next 之前。
		if cursor.next == nil || version > cursor.next.version {

			// 插入到 cursor->next 之前
			next := cursor.next

			// WTF are we spinning on this?
			// 忽略已删除元素
			if next != nil && *next.deleted {
				continue
			}

			// 构造新节点
			newNode := &Node{
				version: version,
				deleted: &f,
				next:    next,
				object: data,
			}


			// 原子的将 cursor.next 由 next 更新为 newNode， 如果失败就递归调用 ll.Insert() 函数，相当于不断的重试 (while)
			if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&cursor.next)), unsafe.Pointer(next), unsafe.Pointer(newNode),
			) {
				return ll.Insert(version, data)
			}

			return newNode
		}

		cursor = cursor.next
	}
}

func assignTrue() *bool {
	b := true
	return &b
}

// Delete deletes the shit out bruh
func (ll *LinkedList) Delete(version uint64) {
	var prev *Node
	currentHead := ll.head
	cursor := ll.head
	var t *bool
	t = assignTrue()



	for {


		if cursor == nil {
			break
		}


		if cursor.version == version {


			if !atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&cursor.deleted)), unsafe.Pointer(cursor.deleted), unsafe.Pointer(t)) {
				ll.Delete(version)
				return
			}

			rt := false

			if prev != nil {
				rt = atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&(prev.next))), unsafe.Pointer(prev.next), unsafe.Pointer(cursor.next))
			} else {
				// HEAD!
				rt = atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&currentHead)), unsafe.Pointer(currentHead), unsafe.Pointer(cursor.next))
			}

			if !rt {
				ll.Delete(version)
			}

			break
		}

		prev = cursor
		cursor = cursor.next
	}
}


// Head returns the object stored in the first node
func (ll *LinkedList) Head() unsafe.Pointer {
	return ll.head.object
}

// LatestVersion returns the node that has a version equal to or less than the version given
func (ll *LinkedList) LatestVersion(v uint64) unsafe.Pointer {
	cur := ll.head
	for cur != nil && cur.version > v {
		cur = cur.next
	}

	if cur == nil {
		return nil
	}

	return cur.object
}

// Snapshot gets the current Snapshot.
// For debugging only
func (ll *LinkedList) Snapshot() (s []uint64) {
	cursor := ll.head

	for cursor != nil {
		if !*cursor.deleted {
			s = append(s, cursor.version)
		}

		cursor = cursor.next
	}

	return
}
