package global

type Topics int

const (
	Consult Topics = iota
	Reserve
)

var TopicNames = map[Topics]string{
	Consult: "consult",
	Reserve: "reserve",
}

func (t Topics) String() string {
	return TopicNames[t]
}

type MQTT_Message struct {
	Topic   Topics `json:"topic"`
	Message string `json:"message"`
}
