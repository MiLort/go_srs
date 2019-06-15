package amf0
type SrsAmf0Number struct {
	value float64
}

func NewSrsAmf0Number(data float64) *SrsAmf0Number {
	return &SrsAmf0Number{
		value:data
	}
}

func (this *SrsAmf0Number) Decode(stream *utils.SrsStream) error {
	marker, err := stream.ReadByte()
	if err != nil {
		return err
	}

	if marker != RTMP_AMF0_Number {
		err := errors.New("amf0 check string marker failed.")
		return err
	}

	this.value, err := stream.ReadFloat64(binary.BigEndian)
	if err != nil {
		return err
	}
	return nil
}

func (this *SrsAmf0Number) Encode(stream *utils.SrsStream) error {
	stream.WriteByte(RTMP_AMF0_Number)
	stream.WriteFloat64(this.value)
	return nil
}

func (this *SrsAmf0Null) IsMyType(stream *utils.SrsStream) (bool, error) {
	marker, err := stream.PeekByte()
	if err != nil {
		return err
	}

	if marker != RTMP_AMF0_Number {
		return false, nil
	}
	return true, nil
}