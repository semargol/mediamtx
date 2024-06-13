package rtp

import (
	"fmt"
	"github.com/pion/rtp"
	"sync"
)

const sortedQueuePower int64 = 4096

// An Item is something we manage in a priority queue.
type SortedQueueItem struct {
	packet     *rtp.Packet // The value of the item; arbitrary.
	index      int64
	timeToSend float64
}

// A PriorityQueue implements heap.Interface and holds Items.
type SortedQueue struct {
	items      []SortedQueueItem
	minIndex   int64
	writeIndex int64
	readIndex  int64
	writeSeq   uint16
	mu         sync.Mutex
}

func (pq *SortedQueue) init() {
	pq.items = make([]SortedQueueItem, sortedQueuePower)
	pq.writeIndex = 0
	pq.minIndex = 65536 * 65536
	pq.readIndex = 0
}

func (pq *SortedQueue) Put(p *rtp.Packet, tts float64) {
	if pq.items == nil {
		pq.init()
	}
	pq.writeSeq = p.SequenceNumber
	if int32(p.SequenceNumber) < int32(pq.writeSeq)-512 {
		wi := ((pq.writeIndex/65536)+1)*65536 + int64(p.SequenceNumber)
		if pq.writeIndex < wi {
			pq.writeIndex = wi
		}
	} else {
		wi := (pq.writeIndex/65536)*65536 + int64(p.SequenceNumber)
		if pq.writeIndex < wi {
			pq.writeIndex = wi
		}
	}
	if pq.minIndex > pq.writeIndex {
		pq.minIndex = pq.writeIndex
	}
	nb, _ := p.Marshal()
	np := new(rtp.Packet)
	_ = np.Unmarshal(nb)
	pq.items[pq.writeIndex%sortedQueuePower] = SortedQueueItem{np, pq.writeIndex, tts}
}

func (pq *SortedQueue) FirstAfter(start int64, finish int64, maxts float64) *SortedQueueItem {
	var mints float64 = maxts
	var found *SortedQueueItem = nil
	for index := start; index <= finish; index++ {
		pi := &pq.items[pq.readIndex%sortedQueuePower]
		if mints > pi.timeToSend {
			mints = pi.timeToSend
			found = pi
		}
	}
	return found
}

func (pq *SortedQueue) Count() int {
	if pq.minIndex > pq.writeIndex {
		return 0
	} else if pq.readIndex < pq.minIndex {
		pq.readIndex = pq.minIndex
	}
	if pq.readIndex < pq.writeIndex-sortedQueuePower+1 {
		pq.readIndex = pq.writeIndex - sortedQueuePower + 1
	}
	return int(pq.writeIndex - pq.readIndex)
}

func (pq *SortedQueue) Get(utc float64) (p *rtp.Packet, tts float64) {
	if pq.minIndex > pq.writeIndex {
		return nil, 0
	} else if pq.readIndex < pq.minIndex {
		pq.readIndex = pq.minIndex
	}
	if pq.readIndex < pq.writeIndex-sortedQueuePower+1 {
		pq.readIndex = pq.writeIndex - sortedQueuePower + 1
	}
	for {
		if pq.readIndex >= pq.writeIndex {
			return nil, 0
		}
		//ri := pq.readIndex % sortedQueuePower
		pi := &pq.items[pq.readIndex%sortedQueuePower]
		if pi.index == pq.readIndex {
			if pi.timeToSend <= utc {
				rep_p := pi.packet
				rep_t := pi.timeToSend
				pi.packet = nil
				pi.index = 0
				pi.timeToSend = 0
				pq.readIndex++
				return rep_p, rep_t
			} else {
				return nil, pi.timeToSend
			}
		}
		fmt.Printf("missed packet number %5d %5d\n", pq.readIndex/65536, pq.readIndex%65536)
		pq.readIndex++
	}
}
