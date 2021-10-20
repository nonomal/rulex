package stdlib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"regexp"
	"rulex/typex"
	"strconv"
	"strings"

	"github.com/ngaut/log"
	lua "github.com/yuin/gopher-lua"
)

var regexper *regexp.Regexp
var pattern = `[a-z]+:[1-9]+`

func init() {
	regexper = regexp.MustCompile(pattern)

}

type BinaryLib struct {
	regexper *regexp.Regexp
}

func NewBinaryLib() typex.XLib {
	return &BinaryLib{
		regexper: regexp.MustCompile(pattern),
	}
}
func (l *BinaryLib) Name() string {
	return "MatchBinary"
}
func (l *BinaryLib) LibFun(rx typex.RuleX) func(*lua.LState) int {
	return func(l *lua.LState) int {
		expr := l.ToString(2)
		data := l.ToString(3)
		returnMore := l.ToBool(4)
		log.Debug("BinaryLib:", expr, data, returnMore)
		// DataToMongo(rx, id, data)
		// Match(expr, []byte(data), returnMore)
		return 0
	}
}

//------------------------------------------------------------------------------------
// 自定义实现函数
//------------------------------------------------------------------------------------

//
// 从一个字节里面提取某 1 个位的值，只有 0 1 两个值
//

func GetABitOnByte(b byte, position uint8) (v uint8, errs error) {
	//  --------------->
	//  7 6 5 4 3 2 1 0
	// |.|.|.|.|.|.|.|.|
	//
	mask := 0b00000001
	if position == 0 {
		return (b & byte(mask)) >> position, nil
	} else {
		return (b & (1 << mask)) >> position, nil
	}
}

//
// TODO: 下一个大版本支持，至少3个月后
//
// 这里借鉴了下Erlang的二进制语法: <<A:5,B:4>> = <<"helloworld">>
// 其中A = hello B= world
//

func ByteToBitFormatString(b []byte) string {
	s := ""
	for _, v := range b {
		log.Debug(v)
		s += fmt.Sprintf("%08b", v)
	}
	return s
}

type Kl struct {
	K  string      //Key
	L  uint        //Length
	BS interface{} //BitString
}

func (k Kl) String() string {
	return fmt.Sprintf("KL@ K: %v,L: %v,BS: %v", k.K, k.L, k.BS)
}

//
// Big-Endian:  高位字节放在内存的低地址端，低位字节放在内存的高地址端。
// Little-Endian: 低位字节放在内存的低地址段，高位字节放在内存的高地址端
//
func ByteToInt(b []byte, order binary.ByteOrder) uint64 {
	log.Debug(b)
	log.Debugf("%08b\n", b)
	var err error
	// uint8
	if len(b) == 1 {
		var x uint8
		err = binary.Read(bytes.NewBuffer(b), order, &x)
		if err != nil {
			log.Error(b, err)
		}
		return uint64(x)
	}
	// uint16
	if len(b) > 1 && len(b) <= 2 {
		var x uint16
		err = binary.Read(bytes.NewBuffer(b), order, &x)
		if err != nil {
			log.Error(b, err)
		}
		return uint64(x)
	}
	// uint32
	if len(b) > 2 && len(b) <= 4 {
		var x uint32
		err = binary.Read(bytes.NewBuffer(b), order, &x)
		if err != nil {
			log.Error(b, err)
		}
		return uint64(x)
	}
	// uint64
	if len(b) > 4 && len(b) <= 8 {
		var x uint64
		err = binary.Read(bytes.NewBuffer(b), order, &x)
		if err != nil {
			log.Error(b, err)
		}
		return x
	}

	return 0
}
func BitStringToBytes(s string) ([]byte, error) {
	b := make([]byte, (len(s)+(8-1))/8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '1' {
			return nil, errors.New("value out of range")
		}
		b[i>>3] |= (c - '0') << uint(7-i&7)
	}
	return b, nil
}
func endian(endian byte) binary.ByteOrder {
	// < 小端 0x34 0x12
	if endian == '>' {
		return binary.BigEndian
	}
	// > 大端 0x12 0x34
	if endian == '<' {
		return binary.LittleEndian
	}
	return binary.LittleEndian
}

//
//
//
func append0Prefix(n int) string {
	if (n % 8) == 7 {
		return "0"
	}
	if (n % 8) == 6 {
		return "00"
	}
	if (n % 8) == 5 {
		return "000"
	}
	if (n % 8) == 4 {
		return "0000"
	}
	if (n % 8) == 3 {
		return "00000"
	}
	if (n % 8) == 2 {
		return "000000"
	}
	if (n % 8) == 1 {
		return "0000000"
	}
	return ""
}

//--------------------------------------------------------------
// stdlib:Match
//--------------------------------------------------------------

func Match(expr string, data []byte, returnMore bool) []Kl {
	cursor := 0
	result := []Kl{}
	matched, err0 := regexp.MatchString(pattern, expr[1:])
	if matched {
		bfs := ByteToBitFormatString(data)

		// <a:12 b:12
		for _, v := range regexper.FindAllString(expr[1:], -1) {
			kl := strings.Split(v, ":")
			k := kl[0]
			if l, err1 := strconv.Atoi(kl[1]); err1 == nil {
				if cursor+l <= len(bfs) {
					binString := bfs[cursor : cursor+l]
					result = append(result, Kl{
						K:  k,
						L:  uint(l),
						BS: append0Prefix(len(binString)) + binString,
					})
				} else {
					result = append(result, Kl{k, uint(l), nil})
				}
				cursor += l
			}
		}
		if returnMore {
			if cursor < len(bfs) {
				result = append(result, Kl{"_", uint(len(bfs) - cursor), append0Prefix(len(bfs[cursor:])) + bfs[cursor:]})
			}
		}

	} else {
		log.Error(err0)
	}

	return result
}
