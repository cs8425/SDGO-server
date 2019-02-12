package main

import (
//	"encoding/binary"
//	"fmt"
	"sync"
)

type EggPool struct {
	mx         sync.RWMutex
	list       []*EggItem
}

type EggItem struct {
	P    int  // 1% = 1000
	ID   uint16
	C    uint8 // C2~C4
}

func NewEggPool() (*EggPool) {
	return &EggPool{
		list: make([]*EggItem, 0, 6),
	}
}

func (e *EggPool) Add() {

}

func (e *EggPool) GetOne() {

}

