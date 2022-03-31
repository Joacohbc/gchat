# Gchat

Un chat grupal en la CLI escrito de Go con dials (sockets), el cual funciona con la estructura de cliente-servidor se debe iniciar el servidor y luego cada uno de los clientes se va conectando y desconectado. El servidor permite a los usuarios a conectarse y desconectarse, aunque controla máximo de clientes, cada cliente se identifica con un nombre a la hora de entrar al chat

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

Una vez entre el usuario debe identificarse con un nombre que no esté ya elegido. Adicionalmente en el chat se tiene comandos como "clear", que borra todo los mensajes y "exit-chat" que sala del chat.

Para cerrar el servidor simplemente realizar un simple Ctrl+C (^C)
