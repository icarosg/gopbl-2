broker:
	docker build -t mosquittobroker -f broker/Dockerfile broker/
	docker run  --rm --name mosquittobroker -p 1884:1884 -p 9001:9001 mosquittobroker

iniciarIpiranga:
	docker-compose -f docker-compose-servidor-ipiranga.yml up --build

iniciar22:
	docker-compose -f docker-compose-servidor.yml up --build

iniciarShell:
	docker-compose -f docker-compose-servidor-shell.yml up --build

cliente-1:
	docker-compose -f docker-compose-cliente-veiculo.yml up --build
	docker attach cliente-veiculo

ipiranga-1:
	docker build -t cliente-posto-ipiranga-1 -f Dockerfile-cliente-posto .
	docker run -it \
		--rm \
		--name cliente-posto-ipiranga-1 \
		-e POSTO_ID=feira-de-santana-ipiranga \
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
		-e POSTO_ID=sao-gonçalo-ipiranga \
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
		-e POSTO_ID=serrinha-ipiranga \
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
		-e POSTO_ID=feira-de-santana-22 \
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
		-e POSTO_ID=sao-gonçalo-22 \
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
		-e POSTO_ID=serrinha-22 \
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
		-e POSTO_ID=feira-de-santana-shell \
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
		-e POSTO_ID=sao-gonçalo-shell \
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
		-e POSTO_ID=serrinha-shell \
		-e POSTO_LAT=150.108 \
		-e POSTO_LONG=150.324 \
		-e POSTO_SERVIDOR=Shell \
		-e POSTO_CIDADE="Serrinha" \
		cliente-posto-shell-3
