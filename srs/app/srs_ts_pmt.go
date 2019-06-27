package app

import (
	"fmt"
	"encoding/binary"
	"go_srs/srs/utils"
)

type SrsTsPayloadPMTESInfo struct {
	/*
		This is an 8-bit field specifying the type of program element carried within the packets with the PID
		whose value is specified by the elementary_PID. The values of stream_type are specified in Table 2-29.
	*/
	streamType SrsTsStream //流类型，标志是Video还是Audio还是其他数据，h.264编码对应0x1b，aac编码对应0x0f，mp3编码对应0x03

	const1Value0 int8 //3bits
	/*
		This is a 13-bit field specifying the PID of the Transport Stream packets which carry the associated
		program element
	*/
	elemenaryPID int16 //13bits
	const1Value1 int8  //4bits
	/*
		This is a 12-bit field, the first two bits of which shall be '00'. The remaining 10 bits specify the number
		of bytes of the descriptors of the associated program element immediately following the ES_info_length field.
	*/
	ESInfoLength int16 //12bits
	ESInfo       []byte
}

func NewSrsTsPayloadPMTESInfo(s SrsTsStream, pid int16) *SrsTsPayloadPMTESInfo {
	return &SrsTsPayloadPMTESInfo{
		streamType:s,
		elemenaryPID:pid,
		const1Value0:0x07,
		const1Value1:0x0f,
		ESInfoLength:0,
	}
}

func (this *SrsTsPayloadPMTESInfo) Encode(stream *utils.SrsStream) {
	stream.WriteByte(byte(this.streamType))

	var epid int16 = 0
	epid |= this.elemenaryPID & 0x1fff
	epid |= int16((int32(this.const1Value0) << 13) & 0xE000)
	stream.WriteInt16(epid, binary.BigEndian)

	fmt.Printf("**************epid=%x***************\n", epid)

	var esv int16 = 0
	esv |= this.ESInfoLength & 0x0FFF
	esv |= int16((int32(this.const1Value1) << 12) & 0xF000)
	stream.WriteInt16(esv, binary.BigEndian)
	//todo check length
	if this.ESInfoLength > 0 {
		stream.WriteBytes(this.ESInfo)
	}
}

func (this *SrsTsPayloadPMTESInfo) Size() uint32 {
	return 5 + uint32(this.ESInfoLength)
}

type SrsTsPayloadPMT struct {
	psiHeader *SrsTsPayloadPSI
	/*
		program_number is a 16-bit field. It specifies the program to which the program_map_PID is
		applicable. One program definition shall be carried within only one TS_program_map_section. This implies that a
		program definition is never longer than 1016 (0x3F8). See Informative Annex C for ways to deal with the cases when
		that length is not sufficient. The program_number may be used as a designation for a broadcast channel, for example. By
		describing the different program elements belonging to a program, data from different sources (e.g. sequential events)
		can be concatenated together to form a continuous set of streams using a program_number. For examples of applications
		refer to Annex C.
	*/
	programNumber int16 //频道号码，表示当前的PMT关联到的频道，取值0x0001

	// 1B
	/**
	 * reverved value, must be '1'
	 */
	const1Value0 int8 //2bits
	/*
		This 5-bit field is the version number of the TS_program_map_section. The version number shall be
		incremented by 1 modulo 32 when a change in the information carried within the section occurs. Version number refers
		to the definition of a single program, and therefore to a single section. When the current_next_indicator is set to '1', then
		the version_number shall be that of the currently applicable TS_program_map_section. When the current_next_indicator
		is set to '0', then the version_number shall be that of the next applicable TS_program_map_section.
	*/
	versionNumber int8 //5bits 版本号，固定为00000，如果PAT有变化则版本号加1
	/*
		A 1-bit field, which when set to '1' indicates that the TS_program_map_section sent is
		currently applicable. When the bit is set to '0', it indicates that the TS_program_map_section sent is not yet applicable
		and shall be the next TS_program_map_section to become valid.
	*/
	currentNextIndicator int8 //1bit	固定为1就好，没那么复杂
	/*
		The value of this 8-bit field shall be 0x00
	*/
	sectionNumber int8
	/*
		The value of this 8-bit field shall be 0x00.
	*/
	lastSectionNumber 	int8
	const1Value1		int8 //3bits
	/*
		This is a 13-bit field indicating the PID of the Transport Stream packets which shall contain the PCR fields
		valid for the program specified by program_number. If no PCR is associated with a program definition for private
		streams, then this field shall take the value of 0x1FFF. Refer to the semantic definition of PCR in 2.4.3.5 and Table 2-3
		for restrictions on the choice of PCR_PID value
	*/
	PCR_PID int16
	// 2B
	const1Value2 int8 //4bits
	/*
		This is a 12-bit field, the first two bits of which shall be '00'. The remaining 10 bits specify the
		number of bytes of the descriptors immediately following the program_info_length field.
	*/
	programInfoLength int16
	programDescriptor []byte //the len is programInfoLength

	infoes []*SrsTsPayloadPMTESInfo
}

func NewSrsTsPayloadPMT(p *SrsTsPacket) *SrsTsPayloadPMT {
	return &SrsTsPayloadPMT{
		psiHeader: NewSrsTsPayloadPSI(p),
		const1Value0:0x3,
		const1Value1:0x7,
		const1Value2:0x0f,
		infoes:make([]*SrsTsPayloadPMTESInfo, 0),
	}
}

func (this *SrsTsPayloadPMT) Encode(stream *utils.SrsStream) {
	s := utils.NewSrsStream([]byte{})//4
	this.psiHeader.Encode(s) //5

	// fmt.Println("************psilen=", len(s)+4, "**************")
	s.WriteInt16(this.programNumber, binary.BigEndian)

	var b byte = 0
	b |= byte(this.currentNextIndicator & 0x01)
	b |= byte((this.versionNumber << 1) & 0x3e)
	b |= byte(this.const1Value0 << 6) & 0xC0
	s.WriteByte(b)
	
	s.WriteByte(byte(this.sectionNumber))
	s.WriteByte(byte(this.lastSectionNumber))//5

	var ppv int16 = this.PCR_PID & 0x1FFF
	ppv |= int16((int32(this.const1Value1) << 13) & 0xE000)
	s.WriteInt16(ppv, binary.BigEndian)

	var pilv int16 = this.programInfoLength & 0xFFF
    pilv |= int16((int32(this.const1Value2) << 12) & 0xF000)
	s.WriteInt16(pilv, binary.BigEndian)
	
	fmt.Printf("***************ppv=%x pplv=%x**************\n", ppv, pilv)

	if this.programInfoLength > 0 {
		//todo check length 
		s.WriteBytes(this.programDescriptor)
	}

	for i := 0; i < len(this.infoes); i++ {
		this.infoes[i].Encode(s)//4
	}

	CRC32 := utils.MpegtsCRC32(s.Data()[1:])
	s.WriteInt32(int32(CRC32), binary.BigEndian)//4
	stream.WriteBytes(s.Data())
	if len(stream.Data()) < 188 {
		i := 188 - len(stream.Data())
		for j := 0; j < i; j++ {
			stream.WriteByte(0xff)
		}
	}
}

func (this *SrsTsPayloadPMT) Size() uint32 {
	var il uint32 = 0
	for i := 0; i < len(this.infoes); i++ {
		il += this.infoes[i].Size()
	}
	return 9 + uint32(this.programInfoLength) + il + 4
}

func (this *SrsTsPayloadPMT) Decode(stream *utils.SrsStream) error {
	return nil
}

func CreatePMT(context *SrsTsContext, pmtNumber int16, pmtPid int16, vpid int16, vs SrsTsStream, apid int16, as SrsTsStream) *SrsTsPacket {
	pkt := NewSrsTsPacket()

	pkt.tsHeader.syncByte = SRS_TS_SYNC_BYTE
	pkt.tsHeader.transportErrorIndicator = 0
	pkt.tsHeader.payloadUnitStartIndicator = 1
	pkt.tsHeader.transportPriority = 0
	pkt.tsHeader.PID = SrsTsPid(pmtPid)
	pkt.tsHeader.transportScrambingControl = SrsTsScrambledDisabled
	pkt.tsHeader.adaptationFieldControl = SrsTsAdapationControlPayloadOnly
	pkt.tsHeader.continuityCounter = 0

	pkt.payload = NewSrsTsPayloadPMT(pkt)
	var pmt *SrsTsPayloadPMT = pkt.payload.(*SrsTsPayloadPMT)
	pmt.psiHeader.pointerField = 0
	pmt.psiHeader.tableId = SrsTsPsiTableIdPms
	pmt.psiHeader.sectionSyntaxIndicator = 1
	pmt.psiHeader.const0Value = 0
	pmt.psiHeader.const1Value0 = 0x03 //2bits
	pmt.psiHeader.sectionLength = 0   //calc in size

	pmt.programNumber = pmtNumber
	pmt.const1Value0 = 0x3 //2bits
	pmt.versionNumber = 0
	pmt.currentNextIndicator = 1
	pmt.sectionNumber = 0
	pmt.lastSectionNumber = 0
	pmt.programInfoLength = 0
	if as == SrsTsStreamAudioAAC || as == SrsTsStreamAudioMp3 {
		pmt.PCR_PID = apid
		pmt.infoes = append(pmt.infoes, NewSrsTsPayloadPMTESInfo(as, apid))
	}

	// if h.264 specified, use video to carry pcr.
	if vs == SrsTsStreamVideoH264 {
		pmt.PCR_PID = vpid
		pmt.infoes = append(pmt.infoes, NewSrsTsPayloadPMTESInfo(vs, vpid))
	}
	//calc section length
	pmt.psiHeader.sectionLength = int16(pmt.Size())
	return pkt
}