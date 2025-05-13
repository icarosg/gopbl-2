veiculo:
	docker-compose -f docker-compose-cliente-veiculo.yml up --build
	docker attach cliente-veiculo

clienteVeiculo:
	docker build -t clienteVeiculo -f Dockerfile-cliente-veiculo .
	docker run -it --rm --name clienteVeiculo --network app-net -e PORT=$(PORT) -p $(PORT):8080 clienteVeiculo

clientePosto:
	docker build -t clientePosto -f Dockerfile-cliente-posto .
	docker run -it --rm --name clientePosto --network app-net -e PORT=$(PORT) -p $(PORT):8080 clientePosto

servidor-22:
	docker build -t servidor-22-f Dockerfile-servidor .
	docker run -it --rm --name servidor-22 --network app-net -e PORT=$(PORT) -p $(PORT):8080 servidor-22

broker:
	docker build -t mosquittoBroker -f broker/Dockerfile broker/
	docker run  --rm --name mosquittoBroker -p 1884:1884 -p 9001:9001 mosquittoBroker

clean:
	docker rm -f car-client station-client mosquitto 2>/dev/null || true

iniciarIpiranga:
	docker-compose -f docker-compose-servidor-ipiranga.yml up --build

iniciar22:
	docker-compose -f docker-compose-servidor.yml up --build

iniciarShell:
	docker-compose -f docker-compose-servidor-shell.yml up --build

ipiranga-1:
	docker build -t cliente-posto-ipiranga-1 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-ipiranga-1 \
		-e POSTO_ID=ipiranga-fsa \
		-e POSTO_LAT=-12.345 \
		-e POSTO_LONG=-38.1234 \
		-e POSTO_SERVIDOR=Ipiranga \
		-e POSTO_CIDADE="Feira de Santana" \
		cliente-posto-ipiranga-1

ipiranga-2:
	docker build -t cliente-posto-ipiranga-2 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-ipiranga-2 \
		-e POSTO_ID=ipiranga-songa \
		-e POSTO_LAT=-22.345 \
		-e POSTO_LONG=-48.1234 \
		-e POSTO_SERVIDOR=Ipiranga \
		-e POSTO_CIDADE="São Gonçalo" \
		cliente-posto-ipiranga-2

ipiranga-3:
	docker build -t cliente-posto-ipiranga-3 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-ipiranga-3 \
		-e POSTO_ID=ipiranga-serrinha \
		-e POSTO_LAT=-32.345 \
		-e POSTO_LONG=-58.1234 \
		-e POSTO_SERVIDOR=Ipiranga \
		-e POSTO_CIDADE="Serrinha" \
		cliente-posto-ipiranga-3

server22-1:
	docker build -t cliente-posto-server22-1 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-server22-1 \
		-e POSTO_ID=server22-fsa \
		-e POSTO_LAT=30.108 \
		-e POSTO_LONG=80.324 \
		-e POSTO_SERVIDOR=22 \
		-e POSTO_CIDADE="Feira de Santana" \
		cliente-posto-server22-1

server22-2:
	docker build -t cliente-posto-server22-2 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-server22-2 \
		-e POSTO_ID=server22-songa \
		-e POSTO_LAT=50.108 \
		-e POSTO_LONG=100.324 \
		-e POSTO_SERVIDOR=22 \
		-e POSTO_CIDADE="São Gonçalo" \
		cliente-posto-server22-2

server22-3:
	docker build -t cliente-posto-server22-3 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-server22-3 \
		-e POSTO_ID=server22-serrinha \
		-e POSTO_LAT=80.108 \
		-e POSTO_LONG=130.324 \
		-e POSTO_SERVIDOR=22 \
		-e POSTO_CIDADE="Serrinha" \
		cliente-posto-server22-3

shell-1:
	docker build -t cliente-posto-shell-1 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-shell-1 \
		-e POSTO_ID=shell-fsa \
		-e POSTO_LAT=-80.108 \
		-e POSTO_LONG=-130.324 \
		-e POSTO_SERVIDOR=Shell \
		-e POSTO_CIDADE="Feira de Santana" \
		cliente-posto-shell-1

shell-2:
	docker build -t cliente-posto-shell-2 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-shell-2 \
		-e POSTO_ID=shell-songa \
		-e POSTO_LAT=-150.108 \
		-e POSTO_LONG=-150.324 \
		-e POSTO_SERVIDOR=Shell \
		-e POSTO_CIDADE="São Gonçalo" \
		cliente-posto-shell-2

shell-3:
	docker build -t cliente-posto-shell-3 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-shell-3 \
		-e POSTO_ID=shell-serrinha \
		-e POSTO_LAT=150.108 \
		-e POSTO_LONG=150.324 \
		-e POSTO_SERVIDOR=Shell \
		-e POSTO_CIDADE="Serrinha" \
		cliente-posto-shell-3