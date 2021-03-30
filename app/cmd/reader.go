package cmd

import (
	"bytes"
	"context"
	"control-mitsubishi-plc-r-kube/config"
	"control-mitsubishi-plc-r-kube/kanban"
	"encoding/hex"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"sync"

	"control-mitsubishi-plc-r-kube/lib"

	"github.com/pkg/errors"
)

const (
	REQUEST_SUBHEADER  = "5000"
	RESPONSE_SUBHEADER = "D000"

	HOST_STATION_NO_           = "00"
	NETWORK_NO_TO_HOST_STATION = "00"
	PC_NO_TO_HOST_STATION      = "FF"
	TARGET_IO_UNIT_NO          = "FF03"
	BULK_READ_CMD              = "0104"
	SUB_CMD                    = "0000"
	WATCH_TIMER                = "1000"
)

type NisPlcMakerSettings struct {
	settings []*NisPlcMakerSetting `yaml:"settings"`
}

type NisPlcMakerSetting struct {
	Content      string `yaml:"strContent"`
	DataSize     int    `yaml:"iDataSize"`
	DeviceNumber string `yaml:"strDevNo"`
	IO           int    `yaml:"iReadWrite"`
	FlowNumber   int    `yaml:"iFlowNo"`
}

var rcvBufferPool = &sync.Pool{
	New: func() interface{} {
		return bytes.Buffer{}
	},
}

func LoadPlcSettings(cfg *config.Config) (*NisPlcMakerSettings, error) {
	f, err := os.Open(cfg.YamlPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	settings := &NisPlcMakerSettings{}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, settings); err != nil {
		return nil, err
	}

	return settings, nil
}

func ReadCombPlc(ctx context.Context, cfg *config.Config) error {
	pcs, err := LoadPlcSettings(cfg)
	if err != nil {
		return err
	}
	pClient := &PlcClient{}
	client, err := pClient.NewClient(cfg.Server.Addr, cfg.Server.Port)
	if err != nil {
		return err
	}
	errCh := make(chan error, 1)
	go func() {
		initReceiveStream()
		iStartDevNo, err := DoWorkReadStart(client.Conn(), pcs)
		if err != nil {
			errCh <- err
			return
		}
		readBuff := make([]byte, 30)
		readLen, err := client.Read(readBuff)
		resp := readBuff[:readLen]

		// レジスタにデータがなければskip
		if !CheckReadData(resp) {
			log.Print("no start signal get. skip")
			errCh <- err
			return
		}

		resp = resp[11:]
		if !CheckIncReadData(resp, readLen) {
			errCh <- errors.New("CheckIncReadData Err")
			return
		}

		pb, err := SetBytesToPlcPart(resp, pcs, iStartDevNo)
		if err != nil {
			errCh <- err
			return
		}

		recordStart := lib.Byte2Int(pb.RecordingStart)
		if recordStart == 1 {
			kanbanReq := map[string]interface{}{
				"status": 0,
			}

			err = kanban.WriteKanban(ctx, kanbanReq)
			if err != nil {
				errCh <- err
			}
			return
		}

		recordStop := lib.Byte2Int(pb.RecordingStop)
		if recordStop == 1 {
			kanbanReq := map[string]interface{}{
				"status": 1,
			}

			err = kanban.WriteKanban(ctx, kanbanReq)
			if err != nil {
				errCh <- err
			}
			return
		}
		//psr := &PlcStartRec{}
		//psr.SetReadDataStartRec(pb)
		//kanbanReq := map[string]interface{}{
		//	"Serial":         psr.Serial,
		//	"Variety":        psr.Variety,
		//	"ProductNumber":  psr.ProductNumber,
		//	"MoveDirection":  psr.MoveDirection,
		//	"MovePart":       psr.MovePart,
		//	"MoveMechanism":  psr.MoveMechanism,
		//	"InspectionDate": psr.InspectionDate,
		//}

	}()

	select {
	case err := <-errCh:
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
	s := lib.GetBytesFrom32bitWithLE(int64(startDevNum))
	s[3] = byte(0xA8)
	startDev := fmt.Sprintf("%X", s)
	wLen := fmt.Sprintf("%X", lib.GetBytesFrom8bitWithLE(int64(writeLen))[0:2])
	dataLen := fmt.Sprintf("%X", len(WATCH_TIMER+BULK_READ_CMD+SUB_CMD+startDev+wLen))

	return REQUEST_SUBHEADER +
		NETWORK_NO_TO_HOST_STATION +
		PC_NO_TO_HOST_STATION +
		TARGET_IO_UNIT_NO +
		HOST_STATION_NO_ +
		dataLen +
		WATCH_TIMER +
		BULK_READ_CMD +
		SUB_CMD +
		startDev +
		wLen
}

//検査開始要求がレジスタに書き込まれてないか常時チェックする
func DoWorkReadStart(conn *net.TCPConn, pc *NisPlcMakerSettings) (int, error) {
	iStartDevNo := math.MaxInt32
	iEndDevNo := math.MinInt32

	for _, device := range pc.settings {
		iDevNo, err := GetDevNo(device.DeviceNumber)
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
