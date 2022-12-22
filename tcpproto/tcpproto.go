package tcpproto

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func PrepareI32(buf []byte, i32 int32) []byte {
	buf = append(buf, byte(i32>>(3*8)), byte(i32>>(2*8)), byte(i32>>(1*8)), byte(i32>>(0*8)))

	return buf
}

func PrepareBytes(buf []byte, bytes []byte) []byte {
	strLen := int32(len(bytes))
	buf = PrepareI32(buf, strLen)
	buf = append(buf, bytes...)

	return buf
}

func PrepareString(buf []byte, str string) []byte {
	return PrepareBytes(buf, []byte(str))
}

func PrepareStrings(buf []byte, strs ...string) []byte {
	arrLen := int32(len(strs))
	buf = PrepareI32(buf, arrLen)
	for _, str := range strs {
		buf = PrepareString(buf, str)
	}

	return buf
}

func WriteMsg(w io.Writer, buf []byte) error {
	fmt.Println("writing msg")
	_, err := w.Write(PrepareI32([]byte{}, int32(len(buf))))
	if err != nil {
		return errors.Wrap(err, "writing msg len: ")
	}
	fmt.Println("writing preps")

	_, err = w.Write(buf)
	if err != nil {
		return errors.Wrap(err, "writing strings: ")
	}
	fmt.Println("writing bufs")

	return nil
}

func ParseI32(buf []byte, i32 *int32) ([]byte, error) {
	if len(buf) < 4 {
		return buf, fmt.Errorf("not enough data for i32")
	}
	*i32 = 0
	*i32 |= int32(buf[0]) << (3 * 8)
	*i32 |= int32(buf[1]) << (2 * 8)
	*i32 |= int32(buf[2]) << (1 * 8)
	*i32 |= int32(buf[3]) << (0 * 8)

	return buf[4:], nil
}

func ParseBytes(buf []byte, bytes *[]byte) ([]byte, error) {
	var bytesLen int32
	buf, err := ParseI32(buf, &bytesLen)
	fmt.Println("bytes recv", len(buf))
	if err != nil {
		return buf, err
	}

	if len(buf) < int(bytesLen) {
		return buf, fmt.Errorf("not enough data for bytes of len %d", bytesLen)
	}

	*bytes = buf[:bytesLen]

	return buf[bytesLen:], nil
}

func ParseString(buf []byte, str *string) ([]byte, error) {
	var strLen int32
	buf, err := ParseI32(buf, &strLen)
	if err != nil {
		return buf, err
	}

	if len(buf) < int(strLen) {
		return buf, fmt.Errorf("not enough data for string of len %d", strLen)
	}

	*str = string(buf[:strLen])

	return buf[strLen:], nil
}

func ParseStrings(buf []byte, strs ...*string) ([]byte, error) {
	var arrLen int32
	buf, err := ParseI32(buf, &arrLen)
	if err != nil {
		return buf, err
	}

	for _, str := range strs {
		buf, err = ParseString(buf, str)
		if err != nil {
			return buf, err
		}
	}

	return buf, nil
}

func ReadMsg(r io.Reader, parse func(buf []byte) error) error {
	msgLenBuf := make([]byte, 4)
	_, err := io.ReadFull(r, msgLenBuf)
	if err != nil {
		return fmt.Errorf("failed to read msg len: %w", err)
	}
	var msgLen int32
	_, err = ParseI32(msgLenBuf, &msgLen)
	if err != nil {
		return fmt.Errorf("failed to parse msg len: %w", err)
	}

	msgBuf := make([]byte, msgLen)
	_, err = io.ReadFull(r, msgBuf)
	if err != nil {
		return fmt.Errorf("failed to read msg: %w", err)
	}

	fmt.Println("recv msg", msgLenBuf, len(msgBuf), msgBuf[:16])

	err = parse(msgBuf)
	if err != nil {
		return fmt.Errorf("failed to parse msg: %w", err)
	}

	return nil
}
