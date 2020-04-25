package packet

import (
	"github.com/pkg/errors"
)

type Packet struct {
	OpCode int16
	Index  int16
	Size   int16
	Data   []byte
}

func (p *Packet) New(opCode int16) {
	p.OpCode = opCode
	p.Data = make([]byte, 0)
}

func (p *Packet) Fill(data []byte) {
	p.Size = (int16)((int16(data[0]&0xFF) << 8) | int16(data[1]&0xFF))
	p.OpCode = (int16)((int16(data[2]&0xFF) << 8) | int16(data[3]&0xFF))

	p.Data = make([]byte, 0)

	for i := 4; i < len(data); i++ {
		p.Data = append(p.Data, data[i])
	}
}

func (p *Packet) Resume(data []byte) {
	for i := 0; i < len(data); i++ {
		p.Data = append(p.Data, data[i])
	}
}

func (p *Packet) IsDone() bool {
	if p.Size == int16(len(p.Data)) {
		return true
	} else {
		return false
	}
}

func (p *Packet) GetBytes() []byte {
	size := len(p.Data)

	var ret []byte

	ret = append(ret, (byte)((size>>8)&0xFF))
	ret = append(ret, (byte)(size&0xFF))
	ret = append(ret, (byte)((p.OpCode>>8)&0xFF))
	ret = append(ret, (byte)(p.OpCode&0xFF))

	for i := 0; i < len(p.Data); i++ {
		ret = append(ret, p.Data[i])
	}
	return ret
}

func (p *Packet) WriteByte(value byte) {
	p.Data = append(p.Data, (byte)(value&0xFF))
	p.Size++
}

func (p *Packet) WriteShort(value int16) {
	p.Data = append(p.Data, (byte)((value>>8)&0xFF))
	p.Data = append(p.Data, (byte)(value&0xFF))
	p.Size += 2
}

func (p *Packet) WriteInteger(value int) {
	p.Data = append(p.Data, (byte)((value>>24)&0xFF))
	p.Data = append(p.Data, (byte)((value>>16)&0xFF))
	p.Data = append(p.Data, (byte)((value>>8)&0xFF))
	p.Data = append(p.Data, (byte)((value)&0xFF))
	p.Size += 4
}

func (p *Packet) WriteString(value string) {
	b := []byte(value)

	p.Data = append(p.Data, (byte)((len(b)>>8)&0xFF))
	p.Data = append(p.Data, (byte)(len(b)&0xFF))

	for i := 0; i < len(b); i++ {
		p.Data = append(p.Data, (byte)(b[i]&0xFF))
	}

	p.Size += int16(len(b)) + 2
}

func (p *Packet) WriteBytes(value []byte) {
	p.Data = append(p.Data, (byte)((len(value)>>8)&0xFF))
	p.Data = append(p.Data, (byte)(len(value)&0xFF))

	for i := 0; i < len(value); i++ {
		p.Data = append(p.Data, (byte)(value[i]&0xFF))
	}

	p.Size += int16(len(value)) + 2
}

func (p *Packet) ReadByte() (byte, error) {
	if p.Index > int16(len(p.Data)) {
		return 0, errors.New("index out of range")
	}

	ret := (byte)(p.Data[p.Index] & 0xFF)
	p.Index++
	return ret, nil
}

func (p *Packet) ReadShort() (int16, error) {
	if p.Index+1 > int16(len(p.Data)) {
		return 0, errors.New("index out of range")
	}

	ret := (int16)((int16(p.Data[p.Index]&0xFF) << 8) | int16(p.Data[p.Index+1]&0xFF))
	p.Index += 2
	return ret, nil
}

func (p *Packet) ReadInteger() (int, error) {
	if p.Index+3 > int16(len(p.Data)) {
		return 0, errors.New("index out of range")
	}

	ret := (int)((int(p.Data[p.Index]&0xFF) << 24) | (int(p.Data[p.Index+1]&0xFF) << 16) | (int(p.Data[p.Index+2]&0xFF) << 8) | int(p.Data[p.Index+3]&0xFF))
	p.Index += 4
	return ret, nil
}

func (p *Packet) ReadString() (string, error) {
	if p.Index+1 > int16(len(p.Data)) {
		return "", errors.New("index out of range")
	}

	length := (int16)((int16(p.Data[p.Index]&0xFF) << 8) | int16(p.Data[p.Index+1]&0xFF))
	p.Index += 2

	if p.Index+length-1 > int16(len(p.Data)) {
		return "", errors.New("index out of range")
	}

	var bret = make([]byte, 0)

	for i := p.Index; i < p.Index+length; i++ {
		bret = append(bret, (byte)(p.Data[i]&0xFF))
	}

	p.Index += length
	return string(bret), nil
}

func (p *Packet) ReadBytes() ([]byte, error) {
	if p.Index+1 > int16(len(p.Data)) {
		return []byte{}, errors.New("index out of range")
	}

	length := (int16)((int16(p.Data[p.Index]&0xFF) << 8) | int16(p.Data[p.Index+1]&0xFF))
	p.Index += 2

	if p.Index+length-1 > int16(len(p.Data)) {
		return []byte{}, errors.New("index out of range")
	}

	var bret = make([]byte, 0)

	for i := p.Index; i < p.Index+length; i++ {
		bret = append(bret, (byte)(p.Data[i]&0xFF))
	}

	p.Index += length
	return bret, nil
}
