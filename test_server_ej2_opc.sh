#!/bin/bash

TEXTO_PRUEBA="Hello World"
if [[ ! $(docker inspect --format='{{.State.Running}}' server) ]]; then 
  echo "El contenedor no se esta ejecutando"
  exit 1
else
  RESPUESTA=$(docker run -i --rm --name test_ej2_opc --network tp0_default alpine sh -c "echo $TEXTO_PRUEBA | nc -w 1 server 5678" )
  if [ $? -eq 0 ]; then 
    RESPUESTA=$(echo "$RESPUESTA" | tr -d '\r')
    if [ -z "$RESPUESTA" ]; then
      echo "No se recibio respuesta"
      exit 1
    elif [[ "$RESPUESTA" == "$TEXTO_PRUEBA" ]]; then 
      echo -e "Test exitoso.\nTexto enviado: $TEXTO_PRUEBA\nTexto recibido: $RESPUESTA" 
      exit 0 
    else 
      echo -e "Test fallido.\nTexto enviado: $TEXTO_PRUEBA\nTexto recibido: $RESPUESTA"
      exit 1
    fi
  else 
    echo "Fallo la ejecuccion del contenedor de prueba"
    exit 1
  fi 
fi