package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"
)


var EggFallBack = &EggItem{
	ID: 0x3AEE, //105屠殺刃
	C: 4,
}



func init() {
	rand.Seed(time.Now().UnixNano())
}

type EggPool struct {
	mx         sync.RWMutex
	list       []*EggItem
}

type EggItem struct {
	P    int  // 1% = 1000
	ID   uint16
	C    uint8 // C2~C4
}

func (i *EggItem) String() string {
	return fmt.Sprintf("[id] %X, [C] %d, [P] %0.3f", i.ID, i.C, float32(i.P) / 1000.0)
}

func NewEggPool() (*EggPool) {
	return &EggPool{
		list: make([]*EggItem, 0, 6),
	}
}

func (e *EggPool) String() string {
	e.mx.Lock()
	defer e.mx.Unlock()

	str := "{\n"
	for _, item := range e.list {
		str += fmt.Sprintf("%v\n", item)
	}
	str += "}\n"
	return str
}

func (e *EggPool) MarshalJSON() ([]byte, error) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	data := struct{
		List  []*EggItem     `json:"list"`
	}{e.list}

	return json.Marshal(data)
}

func (e *EggPool) UnmarshalJSON(in []byte) error {
	data := struct{
		List  []*EggItem     `json:"list"`
	}{}
	err := json.Unmarshal(in, &data)
	if err != nil {
		return err
	}

	e.mx.Lock()
	e.list = data.List
	e.mx.Unlock()

	return nil
}

func (e *EggPool) Add(id uint16, c uint8, p int) {
	e.mx.Lock()
	defer e.mx.Unlock()

	item := &EggItem{
		ID: id,
		C: c,
		P: p,
	}
	e.list = append(e.list, item)
}

func (e *EggPool) GetOne() (*EggItem) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	num := rand.Intn(100 * 1000)
	tmp := 0
	for _, item := range e.list {
		tmp += item.P
		if num < tmp {
			return item
		}
	}
	return EggFallBack // Error, fallback
}

// PacketID = 53 0A
// len() = 52
func BuildEggPack(it *EggItem, gp uint32, pos uint16) ([]byte) {
	buf := Raw2Byte("53 0a 4a 20 00 00" + 
	"EE 3A 00 00 " + 
	"80 25 00 00 " + 
	"00 00 00 00 5d 87 07 00 00 00 00 00 " + 
	"03 " + 
	"EE 3A 00 00 5d 87 07 00 00 00 00 00 " +
	"0c 00 01 00 00 00 00 00 00 fa 44 00 00")

	binary.LittleEndian.PutUint16(buf[6:8], it.ID)
	binary.LittleEndian.PutUint16(buf[27:29], it.ID)
	buf[26] = it.C
	binary.LittleEndian.PutUint16(buf[39:41], pos)

	// GP
	binary.LittleEndian.PutUint32(buf[10:14], gp)

	return buf
}
