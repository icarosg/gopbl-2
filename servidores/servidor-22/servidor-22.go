package main

import (
	"fmt"
	"net/http"
	"sync"

	//mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
)

// gerenciador de tópicos. quais clientes estão inscritos em quais tópicos
type GerenciadorTopicos struct {
	sync.RWMutex
	inscricao map[string]string // tópico ->  cliente (só haverá 1)
}

var gerenciadorTopicos = GerenciadorTopicos{
	inscricao: make(map[string]string),
}

// função para registrar inscrição via API
func inscreverNoTopico(c *gin.Context) {
	topico := c.PostForm("topic")
	clientID := c.PostForm("clientID")

	gerenciadorTopicos.Lock()
	gerenciadorTopicos.inscricao[topico] = clientID
	gerenciadorTopicos.Unlock()

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s subscribed to %s", clientID, topico)})
}

func main() {
	r := gin.Default()
	r.POST("/subscribe", inscreverNoTopico)

	r.Run(":8080")
}
