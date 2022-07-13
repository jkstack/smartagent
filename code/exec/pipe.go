package exec

import (
	"agent/code/report"
	"bufio"
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"io"
	"strings"

	"github.com/jkstack/anet"
	"github.com/lwch/logging"
)

func log(r io.Reader, logger logging.Logger) {
	reader := bufio.NewReader(r)
	for {
		row, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		row = strings.TrimSuffix(row, "\n")
		logger.Write([]byte(row))
	}
}

func send(r io.Reader, ch chan []byte, reporter *report.Data, name string) {
	for {
		var hdr struct {
			Len   uint32
			Crc32 uint32
		}
		err := binary.Read(r, binary.BigEndian, &hdr)
		if err != nil {
			return
		}
		data := make([]byte, hdr.Len)
		_, err = io.ReadFull(r, data)
		if err != nil {
			return
		}
		reporter.PluginReply(name, uint64(hdr.Len))
		if crc32.ChecksumIEEE(data) != hdr.Crc32 {
			logging.Error("invalid crc32")
			return
		}
		var msg anet.Msg
		err = json.Unmarshal(data, &msg)
		if err != nil {
			logging.Error("invalid message")
			return
		}
		ch <- data
	}
}
