FROM eclipse-mosquitto:latest

WORKDIR /mosquitto

RUN mkdir -p /mosquitto/config

COPY mosquitto.conf /mosquitto/config/mosquitto.conf

EXPOSE 1883 9001

CMD ["mosquitto", "-c", "/mosquitto/config/mosquitto.conf"]