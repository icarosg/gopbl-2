# Instruções para execução via Docker

## Configuração

### Servidor
1. Na máquina que será o servidor:
   ```bash
   # Descubra o IP da máquina
   ip addr show  # ou ifconfig no Linux
   ipconfig      # no Windows
   
   # Execute o container do servidor na raiz do projeto
   docker-compose -f docker-compose-servidor.yml up --build
   ```

### Cliente Veículo
1. Na máquina que será o cliente veículo:
   ```bash
   # Edite o arquivo docker-compose-cliente-veiculo.yml e substitua IP_DO_SERVIDOR pelo IP real da máquina do servidor
   
   # Execute o container na raiz do projeto
   docker-compose -f docker-compose-cliente-veiculo.yml up --build
   ```

### Cliente Posto
1. Na máquina que será o cliente posto:
   ```bash
   # Edite o arquivo docker-compose-cliente-posto.yml e substitua IP_DO_SERVIDOR pelo IP real da máquina do servidor
   
   # Execute o container na raiz do projeto
   docker-compose -f docker-compose-cliente-posto.yml up --build
   ```

## Observações Importantes
1. Certifique-se que a porta 1883 está liberada no firewall do servidor
2. As máquinas precisam estar na mesma rede ou ter conectividade entre si
3. Use o IP da interface de rede correta (geralmente a interface da rede local)
4. Se estiver usando uma rede corporativa, verifique se não há bloqueios de porta