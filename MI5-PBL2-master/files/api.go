package main

type ClientType int // Tipo para identificar o tipo do cliente

const (
	CarClientType     ClientType = iota // Cliente carro
	StationClientType                   // Cliente posto
)

var ClientTypeNames = map[ClientType]string{
	CarClientType:     "car_client_type",
	StationClientType: "station_client_type",
}

type Requests int

const (
	RegisterCar           Requests = iota // Registrar carro, passando ID, coordenadas e nível da bateria: carro -> server
	RegisterStation                       // Registrar estação, passando ID e coordenadas: estação -> server
	GetRecommendedStation                 // Obter a estação recomendada para um carro, passando suas coordenadas: carro -> server
	GetReservedStation                    // Obter a estação reservada para um carro, passando seu ID: carro -> server
	IsStationAvailable                    // Verificar se uma estação está disponível, passando seu ID: server -> estações
	ReserveStation                        // Reservar uma estação, passando o ID do carro e o ID da estação: carro -> server
	StartRecharge
	RechargeComplete // Recarregar um carro, passando o ID do carro e o ID da estação: carro -> server
	GeneratePayment  // Gerar um pagamento, passando o ID do pagamento e o valor: server -> carro
	PayRecharge      // Pagar uma recarga, passando o ID do pagamento: carro -> server(Obs:fazer antes da reserva o pagamento)
	UserLogin
	CarUpdate // Sincronizar o carro modficado no cliente pra o servidor, passando o ID do carro e o nível da bateria e posições : carro -> server
	ListStations
	ListActiveStations
	SelectStation // Selecionar uma estação, passando o ID do carro e o ID da estação: carro -> server
	PaymentHistory
	GetStationInfo
	StationUpdate
	ExitStation
	ExitCar
)

var RequestsNames = map[Requests]string{
	RegisterCar:           "register_car",
	RegisterStation:       "register_station",
	GetRecommendedStation: "get_recommended_station",
	GetReservedStation:    "get_reserved_station",
	IsStationAvailable:    "is_station_available",
	ReserveStation:        "reserve_station",
	StartRecharge:         "start_recharge",
	RechargeComplete:      "recharge_car",
	GeneratePayment:       "generate_payment",
	PayRecharge:           "pay_recharge",
	UserLogin:             "user_login",
	CarUpdate:             "battery_sync",
	ListStations:          "list_stations",
	ListActiveStations:    "list_active_stations",
	SelectStation:         "select_station",
	PaymentHistory:        "payment_history",
	GetStationInfo:        "get_station_info",
	StationUpdate:         "station_update",
	ExitStation:           "exit_station",
	ExitCar:               "exit_car",
}

type ResponseStatus int

const (
	Success ResponseStatus = iota
	Error
	Fatal
)

var ResposeStatusNames = map[ResponseStatus]string{
	Success: "success",
	Error:   "error",
	Fatal:   "fatal",
}

func (r ResponseStatus) String() string {
	return ResposeStatusNames[r]
}

func (r Requests) String() string {
	return RequestsNames[r]
}

func (c ClientType) String() string {
	return ClientTypeNames[c]
}

type Message struct {
	ClientType  ClientType     `json:"client_type,omitempty"`
	Req         Requests       `json:"request,omitempty"`
	Status      ResponseStatus `json:"status,omitempty"`
	Err         string         `json:"err,omitempty"`
	Car         Car            `json:"car,omitempty"`
	Station     Station        `json:"station,omitempty"`
	Payment     Payment        `json:"payment,omitempty"`
	CarList     []Car          `json:"car_list,omitempty"`
	StationList []Station      `json:"station_list,omitempty"`
	PaymentList []Payment      `json:"payment_list,omitempty"`
}
