# Protocolos 

## Capa de comunicación

###  Protocolo 

Se propone un protocolo solicitud-respuesta donde emisor envia frames especificado en siguiente seccion, y receptor responde ACK  con largo del mensaje igual a cero, permitiendo al cliente enviar el proximo paquete. 

### Frame 

 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| largo del mensaje               | Mensaje 
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

## Capa de Dominio

### Mensajes enviados 

Los mensajes se envian sobre el frame antes mencionado. Los tipos de mensaje son 

- Agencia: El contenido del frame es el agency_id 

- Apuesta: El contenido del frame son las apuestas separadas por '\n'

- ACK: Frame con largo de mensaje igual a cero. 

### Secuencia de mensajes 

  Cliente                      Servidor

    |                            |
    |--------- Agencia  -------->|  
    |                            |
    |<------ ACK (Response) -----|  
    |                            |
    |--------- Apuesta  -------->|  
    |                            |
    |<------ ACK (Response) -----|  
     (hasta quedarse sin apuestas)

   ( Cliente cierra socket de escritura )  

    |<-------- Apuesta  ---------|  
    |                            |
    |<-------- Apuesta  ---------|  
    |                            |
    |<-------- Apuesta  ---------|  
     (hasta quedarse sin ganadores)

# Concurrencia

Se implemento la solucion por medio de multithreading, donde cada thread maneja un cliente (funcion _handle_client).

La sincronización de Quorum se implementó por medio de la clase SigtermBarrier, presentando mismas funcionalidades que las barreras clasicas pero admitiendo la salida al recibir la señal SIGTERM

Nota: Esta barrera fue inspirada en la implementación de Rust de std::sync::Barrier

# Graceful shutdown 

## Servidor

Graceful shutdown fue implementado por medio de un threading.Event()

Al recibirse la señal SIGTERM se setea dicho evento, haciendo salir al hilo principal y a los hilos de manejo de los clientes. 

## Clientes

Al recibirse SIGTERM se cierra el socket con el servidor, quien descarta al cliente al recibir esta señal. Luego se setea client.state en false, llevando al fin de la rutina sin enviar/recibir mas mensaje. 