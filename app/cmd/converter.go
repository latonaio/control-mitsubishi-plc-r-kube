package cmd

import (
	"bytes"
	"errors"
	"control-mitsubishi-plc-r-kube/lib"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PlcStartRec struct {
	Serial         string    // シリアルNo
	Variety        string    // 品種
	ProductNumber  int       // 品番
	MoveDirection  int       // 動作方向
	MovePart       int       // 動作部位
	MoveMechanism  int       // 動作機構
	DbId           int       // DBインデックス (不要？)
	RecordNo       int       // 録音回数インデックス
	InspectionDate time.Time // 検査日時
}

type PlcStartRecBytes struct {
	Serial        []byte
	Variety       []byte
	ProductNumber []byte
	MoveDirection []byte
	MovePart      []byte
	MoveMechanism []byte
}

// 20数byteのデータを、各パートごとに切り分けていくメソッドにしたい
//https://syoshinsya.mydns.jp/syosinsya/MCprotocol.html
func SetBytesToPlcPart(msg []byte, devices []*NisPlcMakerSetting, startDevNo int) (*PlcStartRecBytes, error) {
	psrb := &PlcStartRecBytes{}
	for _, device := range devices {
		dataSize := device.DataSize * 2

		_, devNo, err := GetDevNo(device.DeviceNumber)
		if err != nil {
			return nil, err
		}

		startIndex := (devNo - startDevNo) * 2

		switch device.Content {
		case "シリアルNo":
			psrb.Serial = bytes.Trim(msg[startIndex:dataSize], "\x00")
		case "品種":
			psrb.Variety = bytes.Trim(msg[startIndex:dataSize], "\x00")
		case "品番":
			psrb.ProductNumber = msg[startIndex:dataSize]
		case "動作方向":
			psrb.MoveDirection = msg[startIndex:dataSize]
		case "動作部位":
			psrb.MovePart = msg[startIndex:dataSize]
		case "動作機構":
			psrb.MoveMechanism = msg[startIndex:dataSize]
		}
	}
	return psrb, nil

}

// 切り分けたbyte列を文字列や数値に直して構造体に入れる
func (psr *PlcStartRec) SetReadDataStartRec(pb *PlcStartRecBytes) {
	psr.Serial = lib.Byte2String(pb.Serial)
	psr.Variety = lib.Byte2String(pb.Variety)
	psr.ProductNumber = lib.Byte2Int(pb.ProductNumber)
	psr.MoveDirection = lib.Byte2Int(pb.MoveDirection)
	psr.MovePart = lib.Byte2Int(pb.MovePart)
	psr.MoveMechanism = lib.Byte2Int(pb.MoveMechanism)
}

//とりあえず移植だけ。リファクタしたほうがいいかも
func GetDevNo(strDevNo string) (strDev string, iDevNo int, err error) {
	cArray := []string{}
	iDevNo = 0
	bFlag := false
	for i, v := range strDevNo {
		sv := string(v)
		isMatch, _ := regexp.MatchString(`^[a-fA-F\\b]+$`, sv)
		if isMatch {
			cArray = append(cArray, sv)
		} else {
			if len(cArray) > 0 {
				str := strDevNo[i : len(strDevNo)-len(cArray)]
				iDevNo, _ = strconv.Atoi(str)
			} else {
				bFlag = true
				return "", 0, errors.New("デバイス番号エラー")
			}
			break
		}
	}
	if !bFlag {
		strDev = strings.Join(cArray, "")
	}
	return strDev, iDevNo, nil
}
