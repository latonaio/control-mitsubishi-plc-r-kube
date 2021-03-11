package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"control-mitsubishi-plc-r-kube/kanban"
	"net"
	"sync"

	"control-mitsubishi-plc-r-kube/lib"

	"github.com/pkg/errors"
)

type NisPlcMakerSetting struct {
	Content      string
	DataSize     int    // データサイズ(word単位)
	DeviceNumber string // デバイス番号
}

var rcvBufferPool = &sync.Pool{
	New: func() interface{} {
		return bytes.Buffer{}
	},
}

func NewNisPlcMakerSetting(content string, dataSize int, deviceNumber string) *NisPlcMakerSetting {
	return &NisPlcMakerSetting{
		Content:      content,
		DataSize:     dataSize,
		DeviceNumber: deviceNumber,
	}
}

func ReadCombPlc(ctx context.Context, targetAddress, targetPort string) error {
	pc := []*NisPlcMakerSetting{
		NewNisPlcMakerSetting("シリアルNo", 16, "D8000"),
		NewNisPlcMakerSetting("品種", 16, "D8020"),
	}
	pClient := &PlcClient{}
	client, err := pClient.NewClient(targetAddress, targetPort)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		initReceiveStream()
		iStartDevNo, err := DoWorkReadStart(client.Conn(), pc)
		if err != nil {
			errCh <- err
		}
		readBuff := make([]byte, 30)
		readLen, err := client.Read(readBuff)
		resp := readBuff[:readLen]

		// 検査開始要求命令が出ていなかったらskip
		if !CheckReadData(resp) {
			log.Print("no start signal get. skip")
			errCh <- err
		}

		resp = resp[11:]
		if !CheckIncReadData(resp, readLen) {
			errCh <- errors.New("CheckIncReadData Err")
		}

		pb, err := SetBytesToPlcPart(resp, pc, iStartDevNo)
		if err != nil {
			errCh <- err
		}

		psr := &PlcStartRec{}
		psr.SetReadDataStartRec(pb)
		kanbanReq := map[string]interface{}{
			"Serial":         psr.Serial,
			"Variety":        psr.Variety,
			"ProductNumber":  psr.ProductNumber,
			"MoveDirection":  psr.MoveDirection,
			"MovePart":       psr.MovePart,
			"MoveMechanism":  psr.MoveMechanism,
			"InspectionDate": psr.InspectionDate,
		}

		err = kanban.WriteKanban(ctx, kanbanReq)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <- errCh:
		return err
	default:
		return nil
	}
}

func initReceiveStream() {
	rcvBuf := rcvBufferPool.Get().(*bytes.Buffer)
	rcvBuf.Reset()
}

// plcにI/O命令を書き込んでる。今回は読み込みだけでOKなのでコマンドは固定
func CreateSendHeader(startDevNum int, writeLen int) string {
	// subheader
	subHeader := "5000"
	//ネットワーク番号
	netNum := fmt.Sprintf("%X", 0)
	//PC番号
	pcNum := fmt.Sprintf("%X", 0xFF)
	//要求先ユニットI/O番号
	io := fmt.Sprintf("%X", 0x3FF)
	//要求先ユニット局番号
	unit := fmt.Sprintf("%X", 0)
	//要求データ長
	dataLen := fmt.Sprintf("%X", 12)
	//CPU監視タイマ
	cpuTimer := fmt.Sprintf("%X", 0x1)
	//コマンド
	cmd := fmt.Sprintf("%X", 0x0401)
	//サブコマンド
	subCmd := fmt.Sprintf("%X", 0x00)
	//要求データ部
	startDev := fmt.Sprintf("%X", startDevNum)
	wLen := fmt.Sprintf("%X", writeLen)

	return subHeader +
		netNum +
		pcNum +
		io +
		unit +
		dataLen +
		cpuTimer +
		cmd +
		subCmd +
		startDev +
		wLen

}

//検査開始要求がレジスタに書き込まれてないか常時チェックする
func DoWorkReadStart(conn *net.TCPConn, pc []*NisPlcMakerSetting) (int, error) {
	iStartDevNo := math.MaxInt32
	iEndDevNo := math.MinInt32

	for _, device := range pc {
		_, iDevNo, err := GetDevNo(device.DeviceNumber)
		if err != nil {
			return -1, err
		}

		if iStartDevNo > iDevNo {
			iStartDevNo = iDevNo
		}

		if iEndDevNo < iDevNo+device.DataSize {
			iEndDevNo = iDevNo + device.DataSize
		}
	}

	tx := CreateSendHeader(iStartDevNo, iEndDevNo-iStartDevNo)
	data, err := hex.DecodeString(tx)
	if err != nil {
		return -1, err
	}
	_, err = conn.Write(data)
	if err != nil {
		return -1, err
	}

	return iStartDevNo, nil
}

func CheckIncReadData(msg []byte, recvLen int) bool {
	idx := 0

	if recvLen < 11 {
		return false
	}

	if msg[idx] != 0xD0 || msg[idx+1] != 0 {
		return false
	}

	idx += 7

	resLen := lib.Byte2Int(msg[idx:2])
	resLen -= 2
	idx += 2

	endCode := lib.Byte2Int(msg[idx:2])
	idx += 2

	if endCode != 0 {
		return false
	}

	return true
}

func CheckReadData(readData []byte) bool {
	if len(readData) > 11 {
		log.Print("Stream Read Err")
		return false
	}
	if readData[0] != 0xD0 || readData[1] != 0 {
		log.Print("Read Header Err")
		return false
	}

	resLength := lib.Byte2Int(readData[7:9]) - 2 //終了コード(2Byte)分差し引いておく
	endCode := lib.Byte2Int(readData[9:11])

	if endCode != 0 {
		log.Print("endCode Err")
		return false
	}
	if resLength < resLength*2 {
		log.Print("ReadSize Err")
		return false
	}

	return true
}
