//
//  queue.go
//
//  Created by Hicham Bouabdallah
//  Copyright (c) 2012 SimpleRocket LLC
//
//  Permission is hereby granted, free of charge, to any person
//  obtaining a copy of this software and associated documentation
//  files (the "Software"), to deal in the Software without
//  restriction, including without limitation the rights to use,
//  copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the
//  Software is furnished to do so, subject to the following
//  conditions:
//
//  The above copyright notice and this permission notice shall be
//  included in all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
//  EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
//  OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
//  HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
//  WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
//  FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
//  OTHER DEALINGS IN THE SOFTWARE.
//

package main

import "sync"

type queuenode struct {
	data string
	id   string
	next *queuenode
}

//	A go-routine safe FIFO (first in first out) data stucture.
type Queue struct {
	head      *queuenode
	tail      *queuenode
	count     int
	lock      *sync.Mutex
	lenOfCmds int
}

//	Creates a new pointer to a new queue.
func NewQueue() *Queue {
	q := &Queue{}
	q.lock = &sync.Mutex{}
	return q
}

//	Returns the number of elements in the queue (i.e. size/length)
//	go-routine safe.
func (q *Queue) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count
}

//	Returns the length of the data (gcode cmd) in the queue (i.e. size/length)
//	go-routine safe.
func (q *Queue) LenOfCmds() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.lenOfCmds
}

//	Pushes/inserts a value at the end/tail of the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Push(item string, id string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := &queuenode{data: item, id: id}

	if q.tail == nil {
		q.tail = n
		q.head = n
	} else {
		q.tail.next = n
		q.tail = n
	}
	q.count++
	q.lenOfCmds += len(item)
}

//	Returns the value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does mutate the queue.
//	go-routine safe.
func (q *Queue) Poll() (string, string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.head == nil {
		return "", ""
	}

	n := q.head
	q.head = n.next

	if q.head == nil {
		q.tail = nil
	}
	q.count--
	q.lenOfCmds -= len(n.data)

	return n.data, n.id
}

//	Returns a read value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does NOT mutate the queue.
//	go-routine safe.
func (q *Queue) Peek() (string, string) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.head
	if n == nil || n.data == "" {
		return "", ""
	}

	return n.data, n.id
}

//	Returns a read value at the front of the queue.
//	i.e. the oldest value in the queue.
//	Note: this function does NOT mutate the queue.
//	go-routine safe.
func (q *Queue) Delete() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.head = nil
	q.tail = nil
	q.count = 0
	q.lenOfCmds = 0
}
