# Pontos-Chave
- Garantir a disponibilidade de todos os postos na viagem
- Permitir planejamento e reserva de múltiplos postos a partir de qualquer servidor
- Utilizar requisições atômicas
- Comunicação entre servidores com MQTT e API REST
- "O cliente que reservar o primeiro ponto deve manter a prioridade na reserva sobre os trechos seguintes,
onde os demais clientes podem desistir ou continuar a compra da passagem escolhendo outros pontos de
carregamento disponíveis" ~Não entendi essa parte
- Pode usar framework
- Deve usar docker, API REST testada com Insominia ou Postman, MQTT e dados gerados aleatoriamente
- Entrega: 12/05

# Problema 
No problema anterior, foi desenvolvido um sistema inteligente de carregamento de veículos elétricos
que pode ser aplicado para gerenciar pontos de recarga em uma cidade. Neste problema, sua startup
identificou a dificuldade dos usuários do sistema em planejar e garantir as recargas necessárias para viagens
longas, entre cidades e estados. Em distâncias longas, é preciso ***garantir a disponibilidade sequencial*** para
completar a viagem dentro de um ***cronograma previsto***, com paradas planejadas de forma ***otimizada e segura***.

O novo desafio da sua equipe é aprimorar o sistema de recarga inteligente para 
***suportar o planejamento e a reserva antecipada de múltiplos pontos de recarga***, dentro de janelas de tempo definidas, ao
longo de uma rota específica entre cidades e estados. O objetivo é que, através de uma ***requisição atômica***, o
sistema possa ***consultar a disponibilidade e reservar*** uma sequência de pontos de recarga necessários para que
o veículo complete sua viagem sem o risco de ficar sem energia, evitando atrasos imprevistos devido à
indisponibilidade de carregadores. 

Para isso, é essencial que exista uma ***forma padronizada e coordenada de comunicação entre os servidores das empresas conveniadas envolvidas***. A ***comunicação entre os servidores*** deve ser realizada ***através de uma API*** projetada pela sua equipe de
desenvolvimento para permitir que um cliente possa, ***a partir de qualquer servidor***, ***reservar pontos*** de
carregamento disponíveis ***em diferentes empresas*** conveniadas seguindo as mesmas regras do sistema
centralizado original. 

Por exemplo, um cliente (carro) que está querendo viajar de João Pessoa à Feira de
Santana pode iniciar a requisição através do servidor da empresa A. Nesta requisição, o cliente escolhe um
ponto de carregamento entre João Pessoa e Maceió, da empresa A, outro ponto de carregamento entre
Maceió e Sergipe, da empresa B, e outro ponto de carregamento entre Sergipe a Feira de Santana, da empresa
C. O ***cliente que reservar o primeiro ponto*** deve manter a ***prioridade na reserva sobre os trechos seguintes***,
onde os ***demais clientes podem desistir ou continuar*** a compra da passagem escolhendo outros pontos de
carregamento disponíveis.

# Restrições
Diferente do anterior, neste problema é ***liberado o uso de frameworks*** de comunicação de terceiros para
implementar a solução do problema, limitados pelos seguintes requisitos:
- Para uma emulação realista do cenário proposto, os elementos da arquitetura devem ser executados
em ***contêineres Docker***, ***executados em computadores distintos*** no laboratório;
- A interface entre os servidores deve ser projetada e implementada através de 
***protocolo baseado em API REST***, podendo ser ***testada*** na apresentação ***através de*** softwares como ***Insomnia ou Postman***;
- Os ***carros*** devem ser ***simulados*** através de um software para geração de dados fictícios, onde os ***dados***
devem ser ***gerados aleatoriamente*** passando a tendência da ***descarga da bateria (rápida, lenta, etc.)***;
- Na comunicação dos carros com o servidor, ***ao invés de uma API de sockets***, estabeleceu-se que a
solução deve adotar o ***padrão usado na Internet das Coisas (IoT)***, com o ***protocolo Message Queue Telemetry Transport (MQTT)***,
classificado como um protocolo ***Machine-to-Machine (M2M)***.

# Cronograma

Entrega: 12/05
Entrega fora do prazo: -20% da nota e -5% por dia de atraso

Apresentação: 12/05 e 14/05

# Avaliação

A nota final será composta por três critérios de avaliação:
1. Desempenho individual (25%)
2. Documentação (25%)
3. Produto Final (código incluso) (50%)