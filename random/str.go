package random

import (
	"math/rand"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bingoohuang/pump/model"
	"github.com/gdamore/encoding"
)

// Str ...
type Str struct {
	maxSize      int
	allowNull    bool
	rr           *RuneRandom
	characterSet string
}

// StrZero ...
func StrZero() reflect.Type {
	return reflect.TypeOf("")
}

// Value ...
// nolint:gomnd
func (r *Str) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < model.NilFrequency {
		return nil
	}

	maxSize := uint64(r.maxSize)

	if maxSize == 0 {
		maxSize = uint64(rand.Int63n(100))
	}

	s := r.rr.Rune(int(maxSize))

	var err error

	// http://technosophos.com/2016/03/09/go-quickly-converting-character-encodings.html
	// https://ipfs.io/ipfs/QmfYeDhGH9bZzihBUDEQbCbTc5k5FZKURMUoUvfmc27BwL/encoding/iso_8859_and_go.html
	// One of the most widely used encodings in the North America and Europe is ISO-8859-1 (aka Latin-1).
	// So here we'll show how to decode from that format into UTF-8.
	// more: https://github.com/bjarneh/latinx, github.com/gdamore/encoding
	// 查看MySQL的表字段字符集
	/*
	   [root@localhost bingoohuang]# ./pump -d "192.168.1.1:3306 test/abc db=sign" -s "show create table signinfo "
	   2020/03/09 23:25:14 SQL: SHOW CREATE TABLE signinfo
	   2020/03/09 23:25:14 cost: 230.097µs
	   +---+----------+---------------------------------------------------------------------------------+
	   | # | TABLE    | CREATE TABLE                                                                    |
	   +---+----------+---------------------------------------------------------------------------------+
	   | 1 | signinfo | CREATE TABLE `signinfo` (                                                       |
	   |   |          |   `id` varchar(64) CHARACTER SET latin1 COLLATE latin1_bin NOT NULL,            |
	   |   |          |   `keyid` varchar(64) CHARACTER SET latin1 COLLATE latin1_bin DEFAULT NULL,     |
	   |   |          |   `signhash` varchar(128) CHARACTER SET latin1 COLLATE latin1_bin DEFAULT NULL, |
	   |   |          |   `signalg` varchar(32) CHARACTER SET latin1 COLLATE latin1_bin DEFAULT NULL,   |
	   |   |          |   `signtime` varchar(64) CHARACTER SET latin1 COLLATE latin1_bin DEFAULT NULL,  |
	   |   |          |   `signdata` varchar(1024) CHARACTER SET utf8 COLLATE utf8_bin DEFAULT NULL,    |
	   |   |          |   `sm2esk` varchar(256) DEFAULT NULL,                                           |
	   |   |          |   `sm2sq` varchar(256) DEFAULT NULL,                                            |
	   |   |          |   `eskgentime` varchar(64) DEFAULT NULL,                                        |
	   |   |          |   PRIMARY KEY (`id`),                                                           |
	   |   |          |   KEY `signtime_index` (`signtime`)                                             |
	   |   |          | ) ENGINE=InnoDB DEFAULT CHARSET=latin1 ROW_FORMAT=COMPACT                       |
	   +---+----------+---------------------------------------------------------------------------------+

	   Mysql系列 —— MySQL的Charset和Collation https://www.itread01.com/content/1532415755.html
	   字符集（character set）：定义了字符以及字符的编码。
	   字符序（collat​​ion）：定义了字符的比较规则。
	*/
	if FoldContains(r.characterSet, "latin1") {
		latin1Encoder := encoding.ISO8859_1.NewEncoder()
		if s, err = latin1Encoder.String(s); err != nil {
			logrus.Panicf("failed to encode to latin1, error %v", err)
		}
	}

	bytes := []byte(s)
	if len(bytes) <= r.maxSize {
		return s
	}

	return string(bytes[0:r.maxSize])
}

// FoldContains tells if s contains sub in fold mode.
func FoldContains(s, sub string) bool {
	us := strings.ToLower(s)
	usub := strings.ToLower(sub)

	return strings.Contains(us, usub)
}

// NewRandomStr ...
func NewRandomStr(column model.TableColumn) *Str {
	return &Str{
		maxSize:      column.GetMaxSize(),
		allowNull:    column.IsNullable(),
		rr:           MakeRuneRandom(),
		characterSet: column.GetCharacterSet(),
	}
}
