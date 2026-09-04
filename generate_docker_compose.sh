#!/bin/bash

if [[ $# -ne 1 ]]; then
	echo "Usage: $0 <cantidad de clientes>"
	exit 1
elif [[ ! $1 =~ ^[+-]?[0-9]*\.?[0-9]+$ ]]; then 
	echo "La cantidad de clientes debe ser numerica y positiva"
	exit 1
elif [[ $1 -lt 1 ]]; then
	echo "La cantidad de clientes debe ser positiva"
	exit 1
fi

echo 'services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    ports:
      - 5678:5678
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678' > docker-compose.yaml

for ((i=0; i<$1; i++)); do echo "
  client_$i:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_$i
    depends_on:
      - server
    environment:
      - AGENCY_ID=$i
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-2.csv
      - OUTPUT_FILE=/output/output-$i.csv
      - BATCH_SIZE=4
    volumes:
      - ./input:/input
      - ./output:/output" >> docker-compose.yaml 
done
