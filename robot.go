package main

import (
	"fmt"
)
/*
機動2/中彈       26 77
機動2/噴氣       28 77
機動3/ZZ        30 77
機動8/降索       31 77
機動8/V降       50 77
機動10/遠射速    4A 77
機動12/衝跳      32 77

攻擊2/hp         04 77
攻擊2/刀圍       05 77
攻擊2/中彈       06 77
攻擊2/遠彈       07 77
攻擊2/噴氣       08 77
攻擊3/9攻        10 77
攻擊6/實彈速     09 77
攻擊6/光判       0A 77
攻擊8/破甲       11 77
攻擊10/刀速      0F 77
攻擊10/中射速    45 77
攻擊10/機動      0E 77
攻擊12/爆擊      12 77
攻擊13/EX       13 77

防禦2/hp          14 77
防禦2/刀圍         15 77
防禦2/中彈         16 77
防禦2/遠彈         17 77
防禦3/9防         20 77
防禦6/光判         1A 77
防禦8/sp升         21 77
防禦10/中射速       47 77
防禦12/最小化       22 77

平衡2/刀圍         35 77
平衡2/中彈         36 77
平衡2/遠彈         37 77
平衡2/噴氣         38 77
平衡3/ZZ          4D 77
平衡6/實彈速       39 77
平衡8/降索         40 77
平衡8/sp升         41 77
平衡10/刀速        3F 77
平衡10/3攻         44 77
平衡12/破甲        42 77
平衡EX/EX         43 77

*/

var (
	WZC = []byte{0x8F, 0x42, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x01, 0x5C, 0x27, 0x30, 0x01, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x71, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC8, 0x44, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE2, 0x07, 0x0C, 0x00, 0x1D, 0x00, 0x10, 0x00, 0x0A, 0x00, 0x3A, 0x00, 0xC0, 0x85, 0x5D, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFD, 0x40, 0x00, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

type Robot struct {
	ID    uint16
	C4    []uint8 // max 4 byte
	Wing  uint8
	WingLv []byte // 4 byte
}

func (r *Robot) GetBytes(pos int) []byte {
	buf := make([]byte, 153, 153)
	copy(buf, WZC)

	buf[0] = byte(r.ID & 0xFF)
	buf[1] = byte((r.ID >> 8) & 0xFF)

	buf[4] = uint8(pos+1) // TODO: mote than 252?

	cc := len(r.C4)
	if cc >= 4 {
		buf[136] = 4
	} else {
		buf[136] = byte(cc)
	}
	switch cc {
	case 4:
		setC(r.C4[3], buf[144:146])
		fallthrough
	case 3:
		setC(r.C4[2], buf[142:144])
		fallthrough
	case 2:
		setC(r.C4[1], buf[140:142])
		fallthrough
	case 1:
		setC(r.C4[0], buf[138:140])

	case 0:
	default:
	}

	buf[148] = r.Wing
	copy(buf[149:153], r.WingLv)

	return buf
}

func setC(cid uint8, buf[]byte) {
	if cid == 0xFF {
		buf[0] = 0x00
		buf[1] = 0x00
	} else {
		buf[0] = cid
		buf[1] = 0x77
	}
}

func NewBot(id uint16) (*Robot) {
	r := &Robot{
		ID: id,
		Wing: 5,
		C4: []uint8{0xFF, 0xFF, 0xFF, 0xFF},
	}

	return r
}

func botPrint(a []byte) {
	wing, wingLV := a[148], a[149:153]
	c1, c2, c3, c4 := a[138:140], a[140:142], a[142:144], a[144:146]
	out := fmt.Sprintf("[id] %X,[pos] %d, [%dc] %X|%X|%X|%X, [w%d] %v", a[0:2], a[4], a[136], c1, c2, c3, c4, wing, wingLV)
	fmt.Println(out)
}


type Grid struct {
	Robot    []*Robot
	buf      [][]byte
}

func NewGrid() (*Grid) {
	return &Grid{
		Robot: make([]*Robot, 0),
		buf: make([][]byte, 0),
	}
}

func (g *Grid) Add(bot *Robot) {
	g.Robot = append(g.Robot, bot)
}

func (g *Grid) BuildCached() {
	g.buf = make([][]byte, 0)

	allbuf := make([][]byte, 0)
	for idx, bot := range g.Robot {
		buf := bot.GetBytes(idx)
		botPrint(buf)
		allbuf = append(allbuf, buf)
	}

	for i := 0; i<len(allbuf); i += 6 {
		pagebuf := make([]byte, 0, 153*6)
		for j := 0; j<6; j++ {
			if i + j >= len(allbuf) {
				break
			}
			pagebuf = append(pagebuf, allbuf[i + j]...)
		}
		g.buf = append(g.buf, pagebuf)
	}
}

func (g *Grid) GetPage(p int) ([]byte) {
	if p < len(g.buf) {
		return g.buf[p]
	}
	return nil
}


