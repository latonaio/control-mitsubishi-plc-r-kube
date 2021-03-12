package cmd

import (
	"control-mitsubishi-plc-r-kube/lib"
	"errors"
	"regexp"
	"strconv"
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
	Serial         []byte
	Variety        []byte
	ProductNumber  []byte
	MoveDirection  []byte
	MovePart       []byte
	MoveMechanism  []byte
	RecordingStart []byte
	RecordingStop  []byte
}

// 20数byteのデータを、各パートごとに切り分けていくメソッドにしたい
//https://syoshinsya.mydns.jp/syosinsya/MCprotocol.html
func SetBytesToPlcPart(msg []byte, devices *NisPlcMakerSettings, startDevNo int) (*PlcStartRecBytes, error) {
	psrb := &PlcStartRecBytes{}
	for _, device := range devices.settings {
		dataSize := device.DataSize * 2

		devNo, err := GetDevNo(device.DeviceNumber)
		if err != nil {
			return nil, err
		}

		startIndex := (devNo - startDevNo) * 2

		switch device.Content {
		case "recording_start":
			psrb.RecordingStart = msg[startIndex:dataSize]
		case "recording_stop":
			psrb.RecordingStart = msg[startIndex:dataSize]
			//case "serial_no":
			//	psrb.Serial = bytes.Trim(msg[startIndex:dataSize], "\x00")
			//case "variety":
			//	psrb.Variety = bytes.Trim(msg[startIndex:dataSize], "\x00")
			//case "product_no":
			//	psrb.ProductNumber = msg[startIndex:dataSize]
			//case "direction_of_movement":
			//	psrb.MoveDirection = msg[startIndex:dataSize]
			//case "part_of_movement":
			//	psrb.MovePart = msg[startIndex:dataSize]
			//case "movement_mechanism":
			//	psrb.MoveMechanism = msg[startIndex:dataSize]
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

//一文字目がアルファベット、二文字目以降が数値という組み合わせ以外であればエラーを返す
func GetDevNo(strDevNo string) (iDevNo int, err error) {
	iDevNo = 0
	m, _ := regexp.MatchString(`^[a-fA-F\\b]+$`, strDevNo[0:1])
	if !m {
		return 0, errors.New("デバイス番号エラー")
	}
	devNo, err := strconv.Atoi(strDevNo[1:])
	if err != nil {
		return 0, errors.New("デバイス番号エラー")
	}

	return devNo, nil
}
