package proto

// Internal representation of proto.Message

type InternalMessage struct {
	Question   Question
	Answer     []ResourceRecord
	Authority  []ResourceRecord
	Additional []ResourceRecord
}

func NewInternalMessage() *InternalMessage {
	return &InternalMessage{}
}

func FromProtoMessage(m Message) *InternalMessage {
	return &InternalMessage{
		Question:   m.Question[0],
		Answer:     m.Answer,
		Authority:  m.Authority,
		Additional: m.Additional,
	}
}

func (im *InternalMessage) ToProtoMessage() Message {
	return Message{
		Header: Header{
			QDCount: 1,
			ANCount: uint16(len(im.Answer)),
			NSCount: uint16(len(im.Authority)),
			ARCount: uint16(len(im.Additional)),
		},
		Question:   []Question{im.Question},
		Answer:     im.Answer,
		Authority:  im.Authority,
		Additional: im.Additional,
	}
}
