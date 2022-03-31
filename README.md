# Gchat

Un chat grupal en la CLI escrito de Go con dials (sockets), el cual funciona con la estructura de cliente-servidor se debe iniciar el servidor y luego cada uno de los clientes se va conectando y desconectado en cualquier momento. Adicionalmente se controlan la cantida máxima de clientes conectadas a la vez. Cada cliente al conectarse debe identificarse con un nombre a la hora de entrar al chat. Nigun mensaje del chat es guardado ni por el servidor ni por los clientes de manera predeterminada.

## Instrucciones

Como primer paso para usar gchat se debe iniciar el servidor. Para iniciarlo se debe ingresar:

```bash
# -p indica el puerto donde se escuchara
# -P el protocolo es "tcp" o "udp" el cual se recibirá
# -u es el límite de usuarios
gchat server -p 8081 -P tcp -u 5
```

```bash
# -p indica el puerto donde se conectara
# -P el protocolo es "tcp" o "udp" en el cual se conectara
gchat client -p 8081 -P tcp
```

Una vez entre el usuario debe identificarse con un nombre que no esté ya elegido. Adicionalmente en el chat se tiene comandos como "clear", que borra todos los mensajes del chat(solo de la pantall del usuario) y "exit-chat" que permite al usuario salirse de la sala chat.

Para cerrar el servidor simplemente realizar un simple Ctrl+C (^C)
