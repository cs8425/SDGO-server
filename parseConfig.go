package main

import (
//	"flag"
//	"log"
	"fmt"

//	"io"
	"os"
	"encoding/hex"
	//"encoding/json"
	"strings"
	"strconv"
	"bufio"
//	"time"
	"sync"
)


// for json encode
type HexByte []byte
func (m HexByte) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%02X"`, m)), nil
}
func (m *HexByte) UnmarshalJSON(in []byte) error {
	buf, err := hex.DecodeString(strings.Trim(string(in), `"`))
	if err != nil {
		return err
	}
	*m = buf
	return nil
}

type HexUint16 uint16
func (m HexUint16) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, m)), nil
}
func (m *HexUint16) UnmarshalJSON(in []byte) error {
	tmp, err :=  strconv.ParseUint(strings.Trim(string(in), `"`), 16, 16)
	if err != nil {
		return err
	}
	*m = HexUint16(tmp)
	return nil
}

type HexUint32 uint32
func (m HexUint32) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, m)), nil
}
func (m *HexUint32) UnmarshalJSON(in []byte) error {
	tmp, err :=  strconv.ParseUint(strings.Trim(string(in), `"`), 16, 32)
	if err != nil {
		return err
	}
	*m = HexUint32(tmp)
	return nil
}

type HexUint64 uint64
func (m HexUint64) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%X"`, m)), nil
}
func (m *HexUint64) UnmarshalJSON(in []byte) error {
	tmp, err :=  strconv.ParseUint(strings.Trim(string(in), `"`), 16, 64)
	if err != nil {
		return err
	}
	*m = HexUint64(tmp)
	return nil
}

type HexColor16 uint16
func (m HexColor16) MarshalJSON() ([]byte, error) {
	out := uint32((m & 0x1F) << 3) // B
	out |= uint32(((m >> 5) & 0x1F) << 3) << 8 // G
	out |= uint32(((m >> 10) & 0x1F) << 3) << 16 // R
	return []byte(fmt.Sprintf(`"%06X"`, out)), nil
}
func (m *HexColor16) UnmarshalJSON(in []byte) error {
	return m.ParseColor(strings.Trim(string(in), `"`), uint16(*m))
}
func (m *HexColor16) ParseColor(str string, def uint16) error {
	tmp, err := strconv.ParseUint(str, 16, 32)
	if err != nil {
		*m = HexColor16(def)
		return err
	}
	out := uint16((tmp & 0xFF) >> 3) // B
	out |= uint16(((tmp >> 8) & 0xFF) >> 3) << 5 // G
	out |= uint16(((tmp >> 16) & 0xFF) >> 3) << 10 // R
	*m = HexColor16(out)
	return nil
}
func (m HexColor16) DumpColor() (out uint32) {
	out = uint32((m & 0x1F) << 3) // B
	out |= uint32(((m >> 5) & 0x1F) << 3) << 8 // G
	out |= uint32(((m >> 10) & 0x1F) << 3) << 16 // R
	return out
}

// for print order
type HexBotID uint16
func (m HexBotID) MarshalJSON() ([]byte, error) {
	id := uint16((m >> 8) | ((m & 0xFF) << 8))
	return []byte(fmt.Sprintf(`"%04X"`, id)), nil
}
func (m *HexBotID) UnmarshalJSON(in []byte) error {
	tmp, err := hex.DecodeString(strings.Trim(string(in), `"`))
	if len(tmp) == 2 && err == nil {
		*m = HexBotID((uint16(tmp[1]) << 8) | uint16(tmp[0]))
	}
	return err
}
func (m HexBotID) String() string {
	id := uint16((m >> 8) | ((m & 0xFF) << 8))
	return fmt.Sprintf("%04X", id)
}



// for parse config only
type UserInfo struct {
	Mx        sync.RWMutex   `json:"-"`

	SearchID  HexBotID
	SearchExp uint32

	Name      string
	GP        uint32
	PageCount int
	GO        int
}

func NewUserInfo() *UserInfo {
	u := &UserInfo{
		SearchID:  0x428F,
		SearchExp: 0,
		GP:        9999999,
		PageCount: 254,
		GO:        1,
	}
	return u
}

func (u *UserInfo) String() string {
	u.Mx.RLock()
	defer u.Mx.RUnlock()
	str := fmt.Sprintf("Name: [% 02X], GP: %d, PageCount: %d, Go: %d, SearchID: %v, SearchExp: %d\n", u.Name, u.GP, u.PageCount, u.GO, u.SearchID, u.SearchExp)
	return str
}

func readData() (error) {
	lines, err := readFile(*userData)
	if err != nil {
		Vln(2, "[open]", err)
		return err
	}

	grid.Clear()

	idx := 1
	for _, line := range lines {

		fields := strings.Split(line, "\t")
		if fields[0] == "" {
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}

		Vln(6, "[dbg]", len(fields), fields)
		if strings.HasPrefix(fields[0], "!!") {
			readUser(fields)
			continue
		}

		var rid uint16 = 0x4286
		var C4 []byte = []byte{0xFF, 0xFF, 0xFF, 0xFF}
		var wing uint8 = 0
		var wingLv []byte = []byte{0x00, 0x00, 0x00, 0x00}
		var Lv uint8 = 13
		var exp uint32 = 12345
		var sess uint32 = 23333
		var skill HexUint32 = 0
		var polish uint16 = 0
		var color []HexColor16 = []HexColor16{0x0000, 0x0000, 0x0000, 0x0000, 0x0000, 0x0000}
		var coat []HexUint32 = []HexUint32{0x00000000, 0x00000000, 0x00000000}


		switch len(fields) {
		default:
			fallthrough
		case 18: // 紋章
			coat[2] = parseUint32LE(fields[17], 16, 0)
			fallthrough
		case 17:
			coat[1] = parseUint32LE(fields[16], 16, 0)
			fallthrough
		case 16:
			coat[0] = parseUint32LE(fields[15], 16, 0)
			fallthrough

		case 15: // 塗裝
			color[5].ParseColor(fields[14], 0)
			fallthrough
		case 14:
			color[4].ParseColor(fields[13], 0)
			fallthrough
		case 13:
			color[3].ParseColor(fields[12], 0)
			fallthrough
		case 12:
			color[2].ParseColor(fields[11], 0)
			fallthrough
		case 11:
			color[1].ParseColor(fields[10], 0)
			fallthrough
		case 10:
			color[0].ParseColor(fields[9], 0)
			fallthrough

		case 9: // 拋光
			tmp, _ := strconv.ParseUint(fields[8], 16, 16)
			polish = uint16(tmp)
			fallthrough
		case 8:
			skill = parseUint32LE(fields[7], 16, 0)
			fallthrough
		case 7:
			tmp, _ := strconv.ParseUint(fields[6], 10, 32)
			sess = uint32(tmp)
			fallthrough
		case 6:
			tmp, _ := strconv.ParseUint(fields[5], 10, 32)
			exp = uint32(tmp)
			fallthrough
		case 5:
			tmp, _ := strconv.ParseUint(fields[4], 10, 8)
			if tmp >= 1 && tmp <= 13 {
				Lv = uint8(tmp)
			}
			fallthrough
		case 4:
			wingLv, _ = hex.DecodeString(fields[3])
			fallthrough
		case 3:
			tmp, _ := strconv.ParseUint(fields[2], 10, 8)
			wing = uint8(tmp)
			fallthrough
		case 2:
			C4, _ = hex.DecodeString(fields[1])
			fallthrough
		case 1:
			tmp, _ := hex.DecodeString(fields[0])
			if len(tmp) == 2 {
				rid = (uint16(tmp[1]) << 8) | uint16(tmp[0])
			}

		case 0:
			Vln(1, "[open]?!!")
			continue
		}

		if rid == 0x0000 {
			grid.DelPos(idx)
			idx += 1
			continue
		}

		bot := NewBot(rid)
		bot.UUID = HexUint64(idx) + 0xADDE0000
		bot.Pos = uint16(idx)
		bot.C4 = C4
		bot.Lv = Lv
		bot.Exp = exp
		bot.Sess = sess
		bot.Wing = wing
		bot.WingLv = wingLv
		bot.Skill = skill
		bot.Polish = polish
		bot.Color = color
		bot.Coat = coat
		bot.Charge = 2000
		bot.C = 4
		//bot.Lock = true

		Vf(5, "[dbg][open]%04X, %04X, %d, %04X, %04X\n", rid, C4, wing, wingLv, color)
		//Vf(7, "[dbg][open]%v, %X", bot, bot.GetBytes(idx))
		idx += 1

		grid.Set(bot)
	}

	grid.SetName(user.Name)
	grid.GP = user.GP
	grid.PageCount = user.PageCount
	grid.SetGoPos(int(user.GO))
	grid.BuildCached()
	grid.BuildCachedAll()

	Vln(4, "[dbg][grid]", len(grid.robot), len(grid.buf), grid.GetGo())
	Vln(4, "[dbg][user]", user)

	return nil
}

func readUser(d []string) {
	if len(d) < 3 {
		return
	}

	val := d[2]
	switch d[1] {
	case "Name":
		//user.SetName(val)
		user.Mx.Lock()
		user.Name = val
		user.Mx.Unlock()

	case "GP":
		tmp, err := strconv.ParseUint(val, 10, 32)
		if err == nil {
			user.Mx.Lock()
			user.GP = uint32(tmp)
			user.Mx.Unlock()
		}

	case "GO":
		tmp, err := strconv.ParseUint(val, 10, 8)
		if err == nil {
			if tmp < 37 && tmp > 0 {
				user.Mx.Lock()
				user.GO = int(tmp)
				user.Mx.Unlock()
			}
		}

	case "SearchID":
		var rid uint16 = 0x4286
		tmp, _ := hex.DecodeString(val)
		if len(tmp) == 2 {
			rid = (uint16(tmp[1]) << 8) | uint16(tmp[0])
		}
		user.Mx.Lock()
		user.SearchID = HexBotID(rid)
		user.Mx.Unlock()

	case "SearchExp":
		tmp, err := strconv.ParseUint(val, 10, 32)
		if err == nil {
			user.Mx.Lock()
			user.SearchExp = uint32(tmp)
			user.Mx.Unlock()
		}

	case "PageCount":
		tmp, err := strconv.ParseUint(val, 10, 32)
		if err == nil {
			user.Mx.Lock()
			user.PageCount = int(tmp)
			user.Mx.Unlock()
		}

	default:
		return
	}
}

func readEggPool() (error) {
	lines, err := readFile(*eggPoolData)
	if err != nil {
		Vln(2, "[open]", err)
		return err
	}

	eggPool2 := NewEggPool()
	for _, line := range lines {

		fields := strings.Split(line, "\t")
		if fields[0] == "" {
			continue
		}
		if strings.HasPrefix(fields[0], "#") {
			continue
		}

		var rid uint16 = 0x4286
		var C uint8 = 3
		var P int = 0

		switch len(fields) {
		default:
			fallthrough
		case 3:
			tmp, _ := strconv.ParseUint(fields[2], 10, 32)
			P = int(tmp)
			fallthrough
		case 2:
			tmp, _ := strconv.ParseUint(fields[1], 10, 8)
			C = uint8(tmp)
			fallthrough
		case 1:
			tmp, _ := hex.DecodeString(fields[0])
			if len(tmp) == 2 {
				rid = (uint16(tmp[1]) << 8) | uint16(tmp[0])
			}

		case 0:
			Vln(1, "[open]?!!")
			continue
		}

		Vf(5, "[dbg][egg][add]%04X, %d, %d\n", rid, C, P)
		eggPool2.Add(rid, C, P)
	}

	eggPool = eggPool2
	Vln(4, "[dbg][eggpool]", eggPool)
	return nil
}

func parseUint32LE(str string, hex int, def uint32) (HexUint32) {
	tmp, err := strconv.ParseUint(str, hex, 32)
	if err != nil {
		return HexUint32(def)
	}
	return HexUint32(tmp)
}

func readFile(path string) ([]string, error) {
	af, err := os.Open(path)
	if err != nil {
		Vln(2, "[open]", err)
		return nil, err
	}
	defer af.Close()

	data := make([]string, 0)
	r := bufio.NewReader(af)
	b, err := r.Peek(3)
	if err != nil {
		return nil, err
	}
	if b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		r.Discard(3)
	}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Trim(line, "\n\r\t")
		data = append(data, line)
	}

	Vln(7, "[dbg][file]", data)
	return data, nil
}

func readExtra() (error) {
	lines, err := readFile(*extraData)
	if err != nil {
		Vln(2, "[extra]", err)
		return err
	}

	tab := make(map[string][]byte)
	tmpName := ""
	tmpVal := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		Vln(6, "[dbg]", line)
		if strings.HasPrefix(line, "$") {
			tmpName = strings.Trim(line, "${ ")
			continue
		}
		if strings.HasPrefix(line, "}") {
			buf := Raw2Byte(tmpVal)
			tmpVal = ""
			if buf == nil {
				Vf(2, "[dbg][extra]%v decode error!!\n", tmpName)
				continue
			}
			tab[tmpName] = buf
			Vf(4, "[dbg][extra]%v = %v[% 02X]\n", tmpName, len(buf), buf)
			continue
		}

		tmpVal += line
	}

	Vln(5, "[dbg][tab]", len(tab), tab)

	// parse
	for k, v := range tab {
		switch k {
		case "UNIT1":
			if len(v) == len(WZC) {
				copy(WZC, v)
				Vf(3, "[extra]update %v[%d]\n", k, len(v))
			}
		case "UNIT2":
			if len(v) == len(IJ) {
				copy(IJ, v)
				Vf(3, "[extra]update %v[%d]\n", k, len(v))
			}
		case "UserInfo001":
			if len(v) == len(UserInfo001) {
				copy(UserInfo001, v)
				Vf(3, "[extra]update %v[%d]\n", k, len(v))
			}
		case "UserInfo002":
			if len(v) == len(UserInfo002) {
				copy(UserInfo002, v)
				Vf(3, "[extra]update %v[%d]\n", k, len(v))
			}
		case "PageFriends":
			if len(v) == len(PageFriends) {
				copy(PageFriends, v)
				Vf(3, "[extra]update %v[%d]\n", k, len(v))
			}
		}
	}

	return nil
}

func dumpRobot() string {
	var out strings.Builder
	list := grid.GetRobotListByPos()
	i := uint16(1)
	for _, b := range list {
		if b.Pos != i {
			out.WriteString("0000\n")
		}
		i = b.Pos + 1
		id := (b.ID >> 8) | ((b.ID & 0xFF) << 8)
		color := fmt.Sprintf("%02X\t%06X\t%06X\t%06X\t%06X\t%06X\t%06X", b.Polish, b.Color[0].DumpColor(), b.Color[1].DumpColor(), b.Color[2].DumpColor(), b.Color[3].DumpColor(), b.Color[4].DumpColor(), b.Color[5].DumpColor())
		coat := fmt.Sprintf("%08X\t%08X\t%08X", b.Coat[0], b.Coat[1], b.Coat[2])
		str := fmt.Sprintf("%04X\t%02X\t%d\t%02X\t%d\t%d\t%d\t%08X\t%v\t%v\n", id, b.C4, b.Wing, b.WingLv, b.Lv, b.Exp, b.Sess, b.Skill, color, coat)
		Vf(5, "[save line]%v", str)
		out.WriteString(str)
	}

	return out.String()
}

func dumpUser() string {
	user.Mx.Lock()
	defer user.Mx.Unlock()

	var out strings.Builder
	out.WriteString("!!!\tName\t")
	out.WriteString(user.Name)

	out.WriteString("\n!!!\tGP\t")
	out.WriteString(fmt.Sprintf("%d", user.GP))

	out.WriteString("\n!!!\tGO\t")
	out.WriteString(fmt.Sprintf("%d", grid.GO))

	out.WriteString("\n!!!\tSearchID\t")
	out.WriteString(fmt.Sprintf("%04X", user.SearchID))
	out.WriteString("\n!!!\tSearchExp\t")
	out.WriteString(fmt.Sprintf("%d\n", user.SearchExp))

	return out.String()
}

func saveData() error {
	botStr := dumpRobot()
	userStr := dumpUser()

	af, err := os.OpenFile(*userData, os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0644)
	if err != nil {
		Vln(2, "[save]", err)
		return err
	}
	defer af.Close()

	_, err = af.WriteString(botStr)
	if err != nil {
		Vln(2, "[save]", err)
		return err
	}
	_, err = af.WriteString(userStr)

	Vln(3, "[saveRobot]")
	return err
}