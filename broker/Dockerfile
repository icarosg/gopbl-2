FROM eclipse-mosquitto:latest

WORKDIR /mosquitto

RUN mkdir -p /mosquitto/config

COPY mosquitto.conf /mosquitto/config/mosquitto.conf

EXPOSE 1884

CMD ["mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]