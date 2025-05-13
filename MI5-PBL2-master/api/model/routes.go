package model

type Route struct {
	ID         string   `bson:"_id,omitempty"` // ID gerado pelo MongoDB
	StartCity  string   `bson:"start_city"`    // Cidade de origem
	EndCity    string   `bson:"end_city"`      // Cidade de destino
	Waypoints  []string `bson:"waypoints"`     // Cidades intermediárias
	Company    string   `bson:"company"`       // Empresa responsável
	DistanceKM int      `bson:"distance_km"`   // Distância total em quilômetros
}

func (r *Route) PrintRoute() {
	println("Route ID:", r.ID)
	println("Start City:", r.StartCity)
	println("End City:", r.EndCity)
	for i, waypoint := range r.Waypoints {
		println("Waypoint", i+1, ":", waypoint)
	}
	println("Company:", r.Company)
	println("Distance (km):", r.DistanceKM)
}
