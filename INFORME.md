# Protocolo de comunicacion 

Se propone una estructura solicitud-respuesta donde los clientes envian el frame de la siguiente seccion, y luego el servidor responde un ACK enviando un frame con largo del mensaje igual a cero, permitiendo al cliente enviar el proximo paquete. 

El servidor es notificado que el cliente termino de enviar apuestas al recibir que el write-end del socket fue cerrado. 

Finalmente el servidor envia u8 largo del mensaje | apuesta 

## Frame 

u8 largo del mensaje | Mensaje 

Mensaje = u8 largo de la agencia (la agencia es string debido a esqueleto de codigo) | agencia | apuesta

# Concurrencia 