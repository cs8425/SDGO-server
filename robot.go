package main

import (
	"encoding/binary"
	//"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"sort"
	"math"
)

var (
	WZC = []byte{0x8F, 0x42, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x01, 0x5C, 0x27, 0x30, 0x01, 0x02, 0x03, 0x04, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x71, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC8, 0x44, 0x29, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE2, 0x07, 0x0C, 0x00, 0x1D, 0x00, 0x10, 0x00, 0x0A, 0x00, 0x3A, 0x00, 0xC0, 0x85, 0x5D, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFD, 0x40, 0x00, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	// len() == 153
	// [0:2] >> ID
	// [4]   >> 位置  uint16

	// [6] 出擊狀態, 0無, 1出擊
	// [7] ???, = 1
	// [8:16] 機體資料庫自動編號(選定出擊用)

	// [16] 等級  Ex == 0x0D (13)
	// [24:25] ????

	// [46:48] 塗裝1
	// [48:50] 塗裝2
	// [50:52] 塗裝3
	// [52:54] 塗裝4
	// [54:56] 塗裝5
	// [56:58] 塗裝6
	// [58:60] 拋光
	// [60:64] 紋章1
	// [64:68] 紋章2
	// [68:72] 紋章3
	// [74:78] 電量 float32LE

	// [78:82]  出戰場數
	// [82:86]  外掛技能?, ID待查

	// [86:102] 取得時間: 年,月,日,時,分,秒,奈秒(4 byte)

	// [132:136] 經驗值
	// [136] >> 特幾
	// [137] >> 上鎻, 0無, 1鎖
	// [138:140] 槽1
	// [140:142] 槽2
	// [142:144] 槽3
	// [144:146] 槽4
	// [148] 合成次數(0~5)
	// [149] 合成亮度
	// [150:153] 合成百分比(直接對應%)

	IJ = []byte{0xC5, 0x3A, 0x00, 0x00, 0x1A, 0x14, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x0C, 0x00, 0x01, 0x01, 0x00, 0x00, 0x00, 0x48, 0x43, 0x00, 0x00}

// len() == 25
// [0:2] >> ID
// [12] 位置
// [15] 技能
// [16] 戰艦
// [17] 上鎖

// [19:23] 電量(float32LE)

// [23] 合成次數(0~5)
// [24] 合成亮度
)

var RobotFallBack = &Robot{
	ID:       0x3AEE,
	Pos:      1,
	UUID:     0xDEAD0001,
	Lock:     false,
	Active:   true,

	C:       4,
	C4:      []uint8{0x4D, 0x30, 0x10, 0x20},
	Wing:    2,
	WingLv:  []byte{0x32, 0x32, 0x32, 0x32},
	Sess:    99999,
	Lv:      13,
	Exp:     0,
	Skill:   0x000124FE,
	Polish:  50,
	Color:   []HexColor16{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000},
	Coat:    []HexUint32{0x00000000, 0x00000000, 0x00000000},
	Charge:  2000,
}

type Robot struct {
	ID     HexBotID
	Pos    uint16
	UUID   HexUint64

	Lock   bool
	Active bool    `json:"-"`

	C      uint8
	C4     HexByte //[]uint8 // max 4 byte
	Wing   uint8
	WingLv HexByte //[]byte // 4 byte
	Sess   uint32
	Lv     uint8
	Exp    uint32
	Skill  HexUint32

	Polish uint16
	Color  []HexColor16 // 6 color
	Coat   []HexUint32 // 3 Coat of Arms

	Charge uint16 // 0~2000, step = 100

	mx    sync.RWMutex // lock for build/read cache
	cache []byte // for packet
}

func (r *Robot) GetBytes() []byte {
	buf := make([]byte, len(WZC), len(WZC))
	copy(buf, WZC)

	buf[0] = byte(r.ID & 0xFF)
	buf[1] = byte((r.ID >> 8) & 0xFF)

	// Pos
	binary.LittleEndian.PutUint16(buf[4:6], r.Pos)

	// UUID
	binary.LittleEndian.PutUint64(buf[8:16], uint64(r.UUID))

	// lock
	if r.Lock {
		buf[137] = 1
	} else {
		buf[137] = 0
	}

	// Active
	if r.Active {
		buf[6] = 1
	} else {
		buf[6] = 0
	}


	cc := uint8(len(r.C4))
	if cc < r.C {
		r.C = cc
	}
	if r.C >= 4 {
		buf[136] = 4
	} else {
		buf[136] = r.C
	}
	switch r.C {
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

	binary.LittleEndian.PutUint32(buf[78:82], r.Sess)

	// LV
	buf[16] = r.Lv

	// Exp
	binary.LittleEndian.PutUint32(buf[132:136], r.Exp)

	// skill
	binary.LittleEndian.PutUint32(buf[82:86], uint32(r.Skill))

	// Polish
	binary.LittleEndian.PutUint16(buf[58:60], r.Polish)

	// Color
	for i, v := range r.Color {
		i = 2 * i
		binary.LittleEndian.PutUint16(buf[46+i:48+i], uint16(v))
	}

	// Coat
	for i, v := range r.Coat {
		i = 4 * i
		binary.LittleEndian.PutUint32(buf[60+i:64+i], uint32(v))
	}

	// Charge
	bits := math.Float32bits(float32(r.Charge))
	binary.LittleEndian.PutUint32(buf[74:78], bits)
	//binary.Write(buf[74:78], binary.LittleEndian, float32(r.Charge))

	return buf
}

func setC(cid uint8, buf []byte) {
	if cid == 0xFF {
		buf[0] = 0x00
		buf[1] = 0x00
	} else {
		buf[0] = cid
		buf[1] = 0x77
	}
}

func (r *Robot) GetBytes2() []byte {
	buf := make([]byte, len(IJ), len(IJ))
	copy(buf, IJ)

	buf[0] = byte(r.ID & 0xFF)
	buf[1] = byte((r.ID >> 8) & 0xFF)

	// Pos
	binary.LittleEndian.PutUint16(buf[12:14], r.Pos)

	buf[23] = r.Wing
	buf[24] = r.WingLv[0]

	// skill
	buf[15] = 0 // r.Skill

	// ship
	buf[16] = 0

	// lock
	if r.Lock {
		buf[17] = 1
	} else {
		buf[17] = 0
	}

	// Charge
	bits := math.Float32bits(float32(r.Charge))
	binary.LittleEndian.PutUint32(buf[19:23], bits)
	//binary.Write(buf[74:78], binary.LittleEndian, float32(r.Charge))

	return buf
}

func NewBot(id uint16) *Robot {
	r := &Robot{
		ID:     HexBotID(id),
		C:      4,
		C4:     []uint8{0xFF, 0xFF, 0xFF, 0xFF},
		Lv:     13,
		Exp:    12345,
		Sess:   0,
		Wing:   0,
		WingLv: []byte{0x00, 0x00, 0x00, 0x00},
	}

	return r
}

func botPrint(a []byte) {
	wing, wingLV := a[148], a[149:153]
	c1, c2, c3, c4 := a[138:140], a[140:142], a[142:144], a[144:146]
	lv := a[16]
	session := binary.LittleEndian.Uint32(a[78:82])
	exp := binary.LittleEndian.Uint32(a[132:136])
	skill := a[82:86]
	pos := binary.LittleEndian.Uint16(a[4:6])
	polish := binary.LittleEndian.Uint16(a[58:60])
	color := fmt.Sprintf("[color]%04X|%04X|%04X|%04X|%04X|%04X", binary.LittleEndian.Uint16(a[46:48]), binary.LittleEndian.Uint16(a[48:50]), binary.LittleEndian.Uint16(a[50:52]), binary.LittleEndian.Uint16(a[52:54]), binary.LittleEndian.Uint16(a[54:56]), binary.LittleEndian.Uint16(a[56:58]))
	out := fmt.Sprintf("[id] %X,[pos] %03d, [%dc] %X|%X|%X|%X, [w%d] %v, [lv] %d, [exp] %d, [sess] %d, [skill] %02X, [Polish] %d%%, %v", a[0:2], pos, a[136], c1, c2, c3, c4, wing, wingLV, lv, exp, session, skill, polish, color)
	fmt.Println(out)
}

func botPrint25B(a []byte) {
	wing, wingLV := a[23], a[24]
	skill := a[15]
	ship := a[16]
	lock := a[17]
	pos := binary.LittleEndian.Uint16(a[12:14])
	out := fmt.Sprintf("[id] %X, [pos] %03d, [w%d] %v, [ship] %d, [skill] %d, [lock] %d", a[0:2], pos, wing, wingLV, ship, skill, lock)
	fmt.Println(out)
}

type Grid struct {
	mx          sync.RWMutex
	robot       map[*Robot]*Robot
	uuid2Robot  map[HexUint64]*Robot
	pos2Robot   map[uint16]*Robot
	GO          uint16

	buf    map[uint16][][]byte
	bufAll []byte // 25 byte * N + header, max(N) = 63

	name      []byte // 1 + 16 Byte
	GP        uint32 // uint32LE
	PageCount int
}

func (g *Grid) MarshalJSON() ([]byte, error) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	data := struct{
		Robot      []*Robot     `json:"list"`
		GO         uint16
		GP         uint32
		PageCount  int
	}{
		Robot: g.GetRobotListByPos(),
		GO: g.GO,
		GP: g.GP,
		PageCount: g.PageCount,
	}

	return json.Marshal(data)
}
func (g *Grid) UnmarshalJSON(in []byte) error {
	data := struct{
		Robot      []*Robot     `json:"list"`
		GO         int
		GP         uint32
		PageCount  int
	}{}
	err := json.Unmarshal(in, &data)
	if err != nil {
		return err
	}
	g.mx.Lock()
	g.GP = data.GP
	g.PageCount = data.PageCount
	g.mx.Unlock()

	g.Clear()
	g.mx.Lock()
	for _, bot := range data.Robot {
		g.add(bot, true)
	}
	g.mx.Unlock()
	g.SetGoPos(data.GO)

	return nil
}


func NewGrid() *Grid {
	g := &Grid{
		GO:     1,

		buf:    make(map[uint16][][]byte),
		bufAll: make([]byte, 0),

		GP:        9999999,
		PageCount: 254,
	}
	g.Clear()
	g.add(RobotFallBack, true)

	return g
}

func (g *Grid) Add(bot *Robot) {
	g.mx.Lock()
	g.add(bot, false)
	g.mx.Unlock()
}

func (g *Grid) Set(bot *Robot) {
	g.mx.Lock()
	g.add(bot, true)
	g.mx.Unlock()
}

func (g *Grid) AddNew(id HexBotID, c uint8) *Robot {
	pos := uint16(0)
	uuid := HexUint64(0xDEAF0000)
	end := uint16(6*g.PageCount)
	for i := uint16(1) ; i < end; i++ {
		_, ok := g.pos2Robot[i]
		if !ok {
			pos = i
			uuid |= HexUint64(i)
			break
		}
	}
	Vf(4, "[AddNew]%02X, %v, %v, %04X\n", id, c, pos, uuid)
	if pos == 0 {
		return nil
	}

	bot := &Robot{
		ID:       id,
		Pos:      pos,
		UUID:     uuid,
		Lock:     false,
		Active:   false,

		C:       c,
		C4:      []uint8{0xFF, 0xFF, 0xFF, 0xFF},
		Wing:    0,
		WingLv:  []byte{0x00, 0x00, 0x00, 0x00},
		Sess:    0,
		Lv:      1,
		Exp:     0,
		Skill:   0,
		Polish:  00,
		Color:   []HexColor16{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000},
		Coat:    []HexUint32{0x00000000, 0x00000000, 0x00000000},
		Charge:  2000,
	}

	g.add(bot, false)

	return bot
}

func (g *Grid) add(bot *Robot, overwrite bool) {
	uuid := bot.UUID
	pos := bot.Pos
	if !overwrite {
		_, ok := g.uuid2Robot[uuid]
		if ok {
			return
		}
		_, ok = g.pos2Robot[pos]
		if ok {
			return
		}
		_, ok = g.robot[bot]
		if ok {
			return
		}
	}
	g.uuid2Robot[uuid] = bot
	g.pos2Robot[pos] = bot
	g.robot[bot] = bot
}

func (g *Grid) del(bot *Robot) {
	uuid := bot.UUID
	pos := bot.Pos

	_, ok := g.uuid2Robot[uuid]
	if !ok {
		return
	}
	_, ok = g.pos2Robot[pos]
	if !ok {
		return
	}
	_, ok = g.robot[bot]
	if !ok {
		return
	}

	delete(g.uuid2Robot, uuid)
	delete(g.pos2Robot, pos)
	delete(g.robot, bot)
}

func (g *Grid) SetPos(bot *Robot, pos int) {
	g.mx.Lock()
	defer g.mx.Unlock()

	bot.Pos = uint16(pos)
	g.add(bot, true)
}

func (g *Grid) DelPos(pos int) {
	g.mx.Lock()
	defer g.mx.Unlock()

	bot, ok := g.pos2Robot[uint16(pos)]
	if !ok {
		return
	}
	g.del(bot)
}

func (g *Grid) Clear() {
	g.mx.Lock()
	defer g.mx.Unlock()

	g.robot = make(map[*Robot]*Robot)
	g.uuid2Robot = make(map[HexUint64]*Robot)
	g.pos2Robot = make(map[uint16]*Robot)
	//g.add(RobotFallBack, true)
}

func (g *Grid) SetGoUUID(uuid uint64) {
	g.mx.Lock()
	defer g.mx.Unlock()

	bot, ok := g.uuid2Robot[HexUint64(uuid)]
	if !ok {
		return // UUID not found
	}

	act, ok := g.pos2Robot[g.GO]
	if !ok {
		return
	}
	act.Active = false

	bot.Active = true
	g.GO = bot.Pos
}

func (g *Grid) SetGoPos(pos int) {
	g.mx.Lock()
	defer g.mx.Unlock()

	bot, ok := g.pos2Robot[uint16(pos)]
	if !ok {
		return // pos not found
	}

	act, ok := g.pos2Robot[g.GO]
	if !ok {
		return
	}
	act.Active = false

	bot.Active = true
	g.GO = uint16(pos)
}

func (g *Grid) GetPos(pos int) *Robot {
	g.mx.RLock()
	defer g.mx.RUnlock()

	bot, ok := g.pos2Robot[uint16(pos)]
	if !ok {
		return nil
	}
	return bot
}

func (g *Grid) GetGo() *Robot {
	g.mx.RLock()
	defer g.mx.RUnlock()

	return g.pos2Robot[g.GO]
}

func (g *Grid) GetRobotList() []*Robot {
	g.mx.RLock()
	defer g.mx.RUnlock()

	list := make([]*Robot, 0, len(grid.robot))
	for _, bot := range grid.robot {
		list = append(list, bot)
	}

	return list
}

type RobotByPos []*Robot
func (r RobotByPos) Len() int      { return len(r) }
func (r RobotByPos) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r RobotByPos) Less(i, j int) bool { return r[i].Pos < r[j].Pos }

func (g *Grid) GetRobotListByPos() []*Robot {
	list := grid.GetRobotList()
	sort.Sort(RobotByPos(list))
	return list
}

func (g *Grid) BuildCached() {
	g.mx.Lock()
	defer g.mx.Unlock()

	allbuf := make(map[uint16][][]byte)
	for _, bot := range g.robot {
		buf := bot.GetBytes()
		botPrint(buf)
		page := (bot.Pos - 1) / 6
		if allbuf[page] == nil {
			allbuf[page] = make([][]byte, 0, 6)
		}
		allbuf[page] = append(allbuf[page], buf)
	}
	g.buf = allbuf
}

func (g *Grid) BuildCachedAll() {
	g.mx.Lock()
	defer g.mx.Unlock()

	size := len(g.robot)
	if size > 63 {
		size = 63
	}
	subbuf := make([]byte, 0, 7+25*size)
	subbuf = append(subbuf, []byte{0xCE, 0x05, 0x85, 0x35, 0x00, 0x00, byte(size)}...)
	count := 0
	for _, bot := range g.robot {
		if count > 63 {
			break
		}
		count += 1

		buf := bot.GetBytes2()
		botPrint25B(buf)
		subbuf = append(subbuf, buf...)
	}
	g.bufAll = subbuf
}

func (g *Grid) GetPage(p int) [][]byte {
	g.mx.RLock()
	defer g.mx.RUnlock()
	return g.buf[uint16(p)]
}

func (g *Grid) GetAllPage() map[uint16][][]byte {
	g.mx.RLock()
	defer g.mx.RUnlock()
	return g.buf
}

func (g *Grid) GetAll() []byte {
	g.mx.RLock()
	defer g.mx.RUnlock()

	return g.bufAll
}

func (g *Grid) String() string {
	g.mx.RLock()
	defer g.mx.RUnlock()
	str := fmt.Sprintf("Name: [% 02X], GP: %d, PageCount: %d, Robot: %d, Go: %04X\n", g.name, g.GP, g.PageCount, len(g.robot), g.GO)
	return str
}

func (u *Grid) SetName(name string) {
	u.mx.Lock()
	defer u.mx.Unlock()

	buf := []byte(name)
	if len(buf) > 16 {
		buf = buf[0:16]
	}

	nameBuf := make([]byte, 17, 17)
	nameBuf[0] = uint8(len(buf))
	copy(nameBuf[1:], buf)

	u.name = nameBuf
}

func (u *Grid) GetInfo1Bytes() []byte {
	u.mx.RLock()
	defer u.mx.RUnlock()

	a := make([]byte, len(UserInfo001), len(UserInfo001))
	copy(a, UserInfo001)

	// Name
	copy(a[8:25], u.name)

	bot := u.GetGo()
	if bot != nil {
		buf := bot.GetBytes()
		copy(a[81:81+len(buf)], buf)
	}

	// GP
	binary.LittleEndian.PutUint32(a[55:59], u.GP)

	return a
}

func (u *Grid) SetPageCount(count int) {
	u.mx.Lock()
	defer u.mx.Unlock()

	if len(u.robot) <= count * 6 {
		u.PageCount = count
	}
}

func (u *Grid) GetPageCountPack() []byte {
	u.mx.RLock()
	defer u.mx.RUnlock()

	buf := Raw2Byte("08 06 85 35 00 00 " +
		"0C 00 " +
		"09 00 F0 03 18 0B 85 35 00 00 03 00 00")

	N := u.PageCount - 4
	if N < 0 {
		N = 0
	}
	N = N * 6

	binary.LittleEndian.PutUint16(buf[6:8], uint16(N))

	return buf
}



var (
	UserInfo001 = Raw2Byte("A6 06 85 35 00 00 00 00 " +
		"10 4A 61 63 6B 30 31 32 33 34 35 36 37 38 39 30 31 " + // [4]"Jack", max = 16 byte
		"00 00 00 00 00 00 00 00 00 00 00 00 00 00 0D 00 00 00 01 00 00 00 15 B9 13 00 86 30 00 00 " +
		"F6 C2 01 00 " + // GP
		"C0 5D 00 00 01 0A 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00 " +
		// "75 42 00 00 20" 出擊機體!!! == 機體資料格式(153 byte)
		"8F 42 00 00 01 00 01 01 19 F3 99 01 00 00 00 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 FA 44 00 00 00 00 00 00 00 00 E3 07 01 00 0E 00 0C 00 1D 00 2C 00 C0 7A F2 04 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 63 0F 03 00 00 00 00 00 00 00 03 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 " + 
		"FA 56 00 00 00 00 00 00 61 EA 00 00 91 76 12 00 01 00 00 01 63 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 E2 07 0C 00 02 00 0C 00 04 00 33 00 40 BD A9 17 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 01 00 00 23 0B 04 00 00 00 00 00 04 00 4E C7 00 00 BF 00 4D 24 14 00 E5 1D 00 00 00 00 00 00 85 35 00 00 B5 1E 04 00 33 C1 1D 00 00 80 D4 44 04 00 00 00 1A 14 88 00 00 00 00 00 B1 3A 00 00 C5 3A 00 00 00 00 00 00 00 00 00 00 87 CC DD 00 00 00 00 00 1A 14 88 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 B1 3A 00 00 11 00 01 01 87 CC DD 00 00 00 00 00 09 00 00 00 00 00 00 00 2C 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 8C 31 80 01 80 01 39 67 01 00 80 65 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 14 00 00 00 00 00 00 00 E2 07 0C 00 14 00 09 00 0D 00 10 00 C0 45 1B 11 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 01 00 EC 09 00 00 03 01 28 77 30 77 26 77 00 00 01 00 00 00 00 00 00 C5 3A 00 00 01 00 01 01 1A 14 88 00 00 00 00 00 0C 00 00 00 00 00 00 00 F1 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 DF 7B E8 18 AF 29 E8 18 5E 57 7F 63 3C 00 6C AA 03 00 65 AA 03 00 63 AA 03 00 00 00 00 00 48 43 7A 00 00 00 00 00 00 00 E2 07 0C 00 0A 00 09 00 2B 00 0B 00 C0 55 70 33 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 01 00 00 00 00 00 03 01 32 77 31 77 30 77 00 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 3A 00 00 00 E8 03 00 00")

// [55:59] GP數量, uint32LE
// [8:25] 玩家名稱, max 16 Byte
// [81:81+153] 出擊機體格式
// [85] 出擊機體位置

	UserInfo002 = Raw2Byte("24 06 7E 08 01 00 " + 
		"03 00 00 00" + // user UUID?
		"10 30 31 32 33 34 35 36 37 38 39 31 32 33 34 35 36 00" + // name
		"08 AE 08 35 19 16 01 F9 03 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 0C 00 00 00 02 00 00 00 " +
		// Exp
		"E9 88 0D 00 50 46 00 00 0C 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 E8 03 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 " +
		// 出擊機體ID
		"8F 42 00 00 00 00 00 00 00 00 00 00 00 00 00 00 0D 00 00 00 00 00 00 00 20 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 FA 44 13 00 00 00 FF 24 01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 F4 AE 08 35 19 16 01 C4 0A AA 34 00 00 10 00 05 42 B3 11 00 01 00 00 00 00 00 00 00 00 00 00 00 00 00 04 00 14 77 20 77 22 77 23 77 01 00 00 00 00 00 00 7C EA 00 00")

// [10:10+17], [10:27] 玩家名稱, max 16 Byte
// [68:72] 官階經驗
// [118:118+153] 出擊機體格式(ID,電量,塗裝,組件,合成)
)

func BuildUserInfo002Pack(name []byte, exp uint32, id HexBotID) []byte {
	a := make([]byte, len(UserInfo002), len(UserInfo002))
	copy(a, UserInfo002)

	// Name
	copy(a[10:27], name)

	// EXP
	binary.LittleEndian.PutUint32(a[68:72], exp)

	// ID
	binary.LittleEndian.PutUint16(a[118:120], uint16(id))

	return a
}
