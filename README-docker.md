# Instruções para execução via Docker

## Pré-requisitos
- Docker instalado
- Docker Compose instalado
- Make instalado (opcional, para usar os comandos do Makefile)

## Configuração Inicial

### Usando Makefile (Recomendado)
O projeto inclui um Makefile que facilita a execução dos containers. Os comandos disponíveis são:

```bash
# Iniciar o broker MQTT
make broker

# Iniciar o servidor (escolha uma das opções)
make iniciarIpiranga    # Para servidor Ipiranga
make iniciar22         # Para servidor 22
make iniciarShell      # Para servidor Shell

# Iniciar cliente veículo
make cliente-1

# Iniciar postos de gasolina (escolha uma das opções)
# Para Ipiranga:
make ipiranga-1        # Posto em Feira de Santana
make ipiranga-2        # Posto em São Gonçalo
make ipiranga-3        # Posto em Serrinha

# Para Servidor 22:
make server22-1        # Posto em Feira de Santana
make server22-2        # Posto em São Gonçalo
make server22-3        # Posto em Serrinha

# Para Shell:
make shell-1           # Posto em Feira de Santana
make shell-2           # Posto em São Gonçalo
make shell-3           # Posto em Serrinha
```

### Configuração Manual

#### 1. Broker MQTT
```bash
# Iniciar o broker MQTT
docker-compose -f docker-compose-broker.yml up -d
```

#### 2. Servidor
1. Na máquina que será o servidor:
   ```bash
   # Descubra o IP da máquina
   ip addr show  # Linux
   ipconfig      # Windows
   
   # Execute o container do servidor 
   docker-compose -f docker-compose-servidor.yml up -d
   ```

#### 3. Cliente Veículo
1. Na máquina que será o cliente veículo:
   ```bash
   # Edite o arquivo docker-compose-cliente-veiculo.yml
   # Substitua IP_DO_SERVIDOR pelo IP real da máquina do servidor
   
   # Execute o container
   docker-compose -f docker-compose-cliente-veiculo.yml up -d
   ```

#### 4. Cliente Posto
1. Na máquina que será o cliente posto:
   ```bash
   # Edite o arquivo docker-compose-cliente-posto.yml
   # Substitua IP_DO_SERVIDOR pelo IP real da máquina do servidor
   
   # Execute o container
   docker-compose -f docker-compose-cliente-posto.yml up -d
   ```

## Acessando os Containers

Para acessar o shell de qualquer container em execução:

```bash
# Listar containers em execução
docker ps

# Acessar o container (substitua CONTAINER_ID pelo ID do container)
docker attach CONTAINER_ID

# Para sair do container sem pará-lo, pressione Ctrl+P seguido de Ctrl+Q
# Para parar o container, pressione Ctrl+C
```



## Observações Importantes
1. Certifique-se que a porta 1883 está liberada no firewall do servidor
2. As máquinas precisam estar na mesma rede ou ter conectividade entre si
3. Use o IP da interface de rede correta (geralmente a interface da rede local)
4. Se estiver usando uma rede corporativa, verifique se não há bloqueios de porta
5. Para parar todos os containers: `docker-compose -f docker-compose-*.yml down`
6. Para remover todos os containers e volumes: `docker-compose -f docker-compose-*.yml down -v`

```

