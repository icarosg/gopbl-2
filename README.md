# Sistema de Monitoramento de Postos de Combustível

Este projeto implementa um sistema distribuído de monitoramento e reserva de postos de combustível, utilizando Go para concorrência e comunicação via MQTT e HTTP. O sistema permite que veículos consultem e reservem postos de combustível em tempo real, com suporte a múltiplas redes de postos (Ipiranga, Shell, etc.).

## Arquitetura do Sistema

### Componentes Principais

1. **Servidores**
   - Gerenciam diferentes redes de postos (Ipiranga, Shell, etc.)
   - Mantêm estado dos postos em banco de dados MongoDB
   - Comunicam-se entre si via HTTP para sincronização
   - Utilizam MQTT para comunicação com clientes
   - Implementam concorrência para gerenciar reservas

2. **Clientes Posto**
   - Simulam postos de combustível individuais
   - Atualizam preços e disponibilidade
   - Comunicam-se com servidores via MQTT
   - Possuem localização geográfica (latitude/longitude)

3. **Clientes Veículo**
   - Consultam postos disponíveis
   - Reservam rotas de postos
   - Recebem atualizações em tempo real
   - Podem escolher servidor preferido

### Comunicação

- **MQTT**: Usado para comunicação em tempo real entre clientes e servidores
- **HTTP**: Usado para comunicação entre servidores
- **MongoDB**: Armazena estado dos postos e suas reservas

## Funcionalidades Detalhadas

### Servidores
- Gerenciamento de postos por rede (Ipiranga, Shell, etc.)
- Sincronização de estado entre servidores
- Processamento de reservas com controle de concorrência
- Atualização de disponibilidade em tempo real
- API REST para comunicação entre servidores
- Integração com MongoDB para persistência

### Clientes Posto
- Simulação de postos com localização geográfica
- Atualização de preços e disponibilidade
- Comunicação assíncrona com servidor via MQTT
- Suporte a diferentes redes de postos
- Configuração via variáveis de ambiente

### Clientes Veículo
- Consulta de postos disponíveis
- Reserva de rotas de postos
- Visualização de preços em tempo real
- Seleção de servidor preferido
- Interface interativa via linha de comando
- Tratamento de concorrência na reserva

## Tecnologias Utilizadas

- **Go**: Linguagem principal do projeto
- **MQTT**: Protocolo de comunicação em tempo real
- **MongoDB**: Banco de dados para persistência
- **Docker**: Containerização dos componentes
- **Docker Compose**: Orquestração dos containers
- **Make**: Automação de tarefas

## Pré-requisitos

- Go 1.16 ou superior
- Docker e Docker Compose
- Make (opcional)
- MongoDB (gerenciado via Docker)

## Estrutura do Projeto

```
.
├── broker/              # Configuração do broker MQTT
├── cliente-posto/       # Implementação dos clientes posto
├── cliente-veiculo/     # Implementação dos clientes veículo
├── servidores/          # Implementação dos servidores
│   ├── servidor-22/     # Servidor da rede 22
│   ├── servidor-Ipiranga/ # Servidor da rede Ipiranga
│   └── Shell/          # Servidor da rede Shell
├── models/             # Modelos de dados compartilhados
├── db/                 # Configurações de banco de dados
└── docker-compose-*.yml # Arquivos de configuração Docker
```

## Execução

### Usando Docker (Recomendado)

Para instruções detalhadas sobre como executar o projeto usando Docker, consulte o arquivo [README-docker.md](README-docker.md).

### Usando Makefile

O projeto inclui um Makefile com comandos úteis:

```bash
# Iniciar o broker MQTT
make broker

# Iniciar servidores (escolha um)
make iniciarIpiranga    # Servidor Ipiranga
make iniciar22         # Servidor 22
make iniciarShell      # Servidor Shell

# Iniciar cliente veículo
make cliente-1

# Iniciar postos (exemplos)
make ipiranga-1        # Posto Ipiranga em Feira de Santana
make server22-1        # Posto Servidor 22 em Feira de Santana
make shell-1           # Posto Shell em Feira de Santana
```

## Detalhes Técnicos

### Concorrência
- Uso de goroutines para operações assíncronas
- Mutex para controle de acesso concorrente
- Canais para comunicação entre goroutines
- Timeouts para operações críticas

### Persistência
- MongoDB para armazenamento de dados
- Coleções separadas por servidor
- Índices para otimização de consultas
- Sincronização entre servidores

### Comunicação
- MQTT para mensagens em tempo real
- HTTP para comunicação entre servidores
- JSON para serialização de dados
- Tópicos MQTT específicos por servidor

## Desenvolvimento

### Dependências Principais
- github.com/eclipse/paho.mqtt.golang
- go.mongodb.org/mongo-driver
- github.com/gin-gonic/gin

### Estrutura de Código
- `models/`: Define estruturas de dados compartilhadas
- `servidores/`: Implementação dos servidores
- `cliente-posto/`: Implementação dos clientes posto
- `cliente-veiculo/`: Implementação dos clientes veículo
- `db/`: Configurações e operações de banco de dados

