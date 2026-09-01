Redactar un breve informe en donde se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado y los mecanismos para sincronizar la ejecución concurrente.

# Protocolo de comunicación
Para lograr el coportamiento enunciado, planteamos el siguiente diagrama de secuencia que de forma burda describe la interacción entre las aplicaciones que en este caso usan la arquitectra cliente-servidor. 
Notando que el cliente en el dominio actua como agencia de lotería que cargan participantes en el sevidor, este último es el responsable de comunicar a quiénes hayan ganado su premio.

<p align="center">
  <img src="img/diagrama1.svg" alt="Ejemplo de diagrama">
</p>

Para eso definimos el protocolo de comunicación de capa de aplicación, teniendo en cuenta de que la capa de transporte usará TCP, tendrá los siguientes mensajes divididos según su funcionalidad:

- Para enviar información del flujo de apuestas
    - BETS: Este mensaje puede contener tantos registros de apuestas como entren en un batch definido en la variable de entorno `BATCH_SIZE`.
    - WINNERS: Este mensaje puede contener tantos registros de apuestas ganadoras de la agencia destino como entren en un batch, puede no contener registros. Además indica el fin de la comunicación entre estos.
    - ERROR: Este mensaje indica que hubo un error en la comunicación, por lo que el cliente debe cerrar la conexión con el servidor.   

- De control
    - ALL_SENDED: Este mensaje indica que el cliente ha enviado todos los registros de apuestas que tenía para enviar.

## Flujo de comunicación
En el siguiente diagrama se muestra el flujo de comunicación esperado, en este caso hay dos clientes.

El cliente/agencia 1, envía varios batches de apuestas (es decir mensajes que contienen tantos rgistros de apuestas como entren en el batch), lo cual lo hace mediante el mensaje **BETS**.
Luego de que haya mandado todos los registros de apuestas, envía el mensaje **ALL_SENDED** para notificar al servidor que ya no enviará más apuestas; es en este punto que el cliente debe quedar a la espera de recibir un mensaje con los ganadores de su agencia o bien vacío si ninguna de sus apuestas fue ganadora.
Análogamente actúa la agencia 2, pero es importante notar que luego de que ambas agencias hayan enviado todas sus apuestas el servidor calcula el ganador.

> NOTA: Para el ejercicio numero 7, se considera que el `AGENCY_QUORUM_MIN` es 2 o bien el servidor solo tiene esas dos conexiones.?
<p align="center">
  <img src="img/ejemplo_protocolo.svg" alt="Protocolo de comunicación">
</p>

En caso de que haya algún tipo de error en la comunicación tanto servidor como cliente deben cerrar la conexión y terminar la ejecución, sea interno o por un mesaje recibido. 
## Estructura de mensajes
En este caso, se implementó la serialización y deserialización de los mensajes enviados por la red que se mapean a un clase directa para poder manipularlos mejor en el código existente de las apuestas.

### Header de el paquete

Todos los mensajes comparten un *header común*, el cual tiene la siguiente estructura:
```
 0                   1                   2                   3   
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     Type      |   Agency ID   |         Payload Length        
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        Payload Length (cont.)  |         Payload ...           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- `Type (1B)`: Indica el tipo de mensaje (BETS, WINNERS, ERROR, ALL_SENDED). 
- `Agency ID (1B)`: Identificador de la agencia que envía el mensaje.
- `Payload Length (4B)`: Indica la longitud del payload en bytes.
- `Payload`: Contiene la información específica del mensaje, como los registros de apuestas o ganadores.

*Nota:* A nivel código, se implementó una clase abstracta `Message` que define la estructura y comportamiento común de todos los mensajes, incluyendo métodos para serializar y deserializar el header y el payload, los cuales varían según el tipo y de ahí que cada uno tenga subclase particular que hereda de `Message` y define su propio comportamiento para el payload. [Ver implementación](services/server/src/server/protocol/messages.py)

### `BETS` payload
Luego, para hacer posible el envío de varios registros de apuestas en un solo mensaje, definimos la estructura del payload para el mensaje BETS de la siguiente manera:
```
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Bet Count (2B)       |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Bet Record 1 (variable)  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|              ...              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|      Bet Record N (variable)  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- `Bet Count (2B)`: Indica la cantidad de registros de apuestas incluidos en el mensaje.
- `Bet Record (variable)`: Cada registro de apuesta tiene una estructura definida que incluye la información de la misma, de la siguiente manera:

```
+----------------+----------------+----------------+----------------+
|                        Document (4B)                              |
+----------------+----------------+----------------+----------------+
|         Number (2B)              | FN Len (1B) | LN Len (1B)      |
+----------------+----------------+----------------+----------------+
|                    Birthdate (8B, "YYYYMMDD")                     |
+-------------------------------------------------------------------+
|              First Name (FN Len bytes, UTF-8)                     |
+-------------------------------------------------------------------+
|              Last Name (LN Len bytes, UTF-8)                      |
+-------------------------------------------------------------------+
```

### `WINNERS` payload
Análogo al mensaje BETS, solo que en este caso el payload contendrá los registros de apuestas ganadoras, con la misma estructura de registro de apuesta definida anteriormente, solo que en este caso no hay límite de batch?

### `ERROR` payload

```
+---------------------------------------------------------------------+
|                    Message (variable, UTF-8)                        |
+---------------------------------------------------------------------+
```
- `Message (variable)`: Contiene un mensaje de error descriptivo en formato UTF-8.

### `ALL_SENDED` y `ACK` payload 
Por un lado, el mensaje `ALL_SENDED` no contiene payload, ya que su función es únicamente indicar que el cliente ha terminado de enviar todos los registros de apuestas.
Mientras que, el mensaje `ACK` no contiene payload, ya que su función es únicamente indicar que el cliente ha recibido correctamente un mensaje del servidor 
para sincronizar las tasas de envío y recepción de mensajes entre el cliente y el servidor, es decir que el cliente puede enviar un nuevo mensaje hasta que haya recibido un ACK del servidor, y viceversa.

## Implementación

Partiendo de las clases en `services/server/src_frozen` que establecen el modelo de dominio de la aplicación, se implementaron clases análogas para el cliente en Go.
Lo que se buscaba es que la aplicación se abstraiga de la serialización y deserialización de los mensajes, para que el cliente pueda enviar y recibir mensajes de manera sencilla, 
es por eso que conectamos las clases que simbolizaban los mensajes con las clases que representaban el modelo de dominio, de manera que el cliente pueda enviar y recibir mensajes sin
tener que serializar o deserializarlos manualmente.

Ejemplo de métodos que se usan para serializar y deserializar los mensajes, sin importar su tipo.

```python
marshall() -> bytes
unmarshall(data: bytes) -> Message
```

Si bien estos mismos no utilizan la clase `Bet`explícitamente, en el caso de que sea un mensaje de tipo `BETS` o `WINNERS`, el payload contendrá registros de apuestas, por lo que se implementaron métodos para  convertir la clase `Bet` a la estructura de registro de apuesta definida en el protocolo que se serializa en el payload, y viceversa.
Esto es posible mediante la abstracción del mensaje `BetMessage` que tiene como atributo una lista de apuestas que se pueden agregar mediante el método `add_bet`, y que luego se serializa en el payload del mensaje.
```python
'''
Agrega un registro de apuesta que se usará para serializar y enviar en un mensaje de tipo `BETS` o `WINNERS`.
Verifica que no se agreguen más de los que deberían entrar en un batch, si es que existe esta limitación.
'''
add_bet(bet: Bet) -> None
```

# Manejo de concurrencia
El servidor soporta múlltiples clientes concurrentes, cada uno se maneja en su propio hilo por la función `handle_client`.

Para que sea seguro que no hayan **race conditions**, se implementó un mecanismo de sincronización para acceder al archivo de apuestas recibidas, que en este se crea con el nombre `bets_received.csv`

## Acceso al archivo de apuestas recibidas

Dado a que el servidor tiene dos operaciones principales sobre el recurso compartido que son de lectura y escritura, 
se implementó un mecanismo de sincronización que permite que múltiples hilos puedan leer el recurso compartido de manera concurrente, pero solo un hilo pueda escribir en él a la vez, y mientras se está escribiendo, ningún otro hilo puede leerlo ni escribirlo.

El mecanismo conocido como Read Write Lock, se implementó en la clase `RWLock` que se encuentra [en este link](services/server/src/server/rw_lock.py), y a grandes rasgos logra lo antes mencionado mediante el uso de una condition variable para evitar hacer busy waiting, 
y un lock para proteger el acceso a las variables que controlan el estado del lock. 

## Sincronización de hilos para alcanzar el quórum
Con la finalidad de que el servidor haga el sorteo respetando el quórumde cantidad mínima de agencias que deberían participar, se usó el macanismo de barrera de hilos de la librería estándar threading de Python (threading.Barrier) 
que permite que un grupo de hilos se bloqueen hasta que todos los hilos del grupo hayan alcanzado un punto de sincronización común, en este caso, el punto de sincronización es cuando la cantidad de agencias que han enviado 
todas sus apuestas es al menos la indicada en la variable de entorno `AGENCY_QUORUMIN`.

Es importante notar que, la barrera libera exactamente `AGENCY_QUORUMIN` clientes para hacer el sorteo.
Ejemplo: en caso de haber inicializado la barrera con `AGENCY_QUORUMIN=3`, el servidor esperará a que 3 agencias hayan enviado todas sus apuestas para hacer el sorteo, y luego de eso, liberará a los 3 clientes para que reciban los resultados de sus apuestas por más de que 
hayan más clientes que hayan enviado todas sus apuestas (es decir estén esperando ser liberados), estos deberán esperar a que se haga otro sorteo para recibir los resultados de sus apuestas.

De esta forma logramos sincronizar la ejecución concurrente de los hilos del servidor, y garantizar que un sorteo se haga solo cuando se haya alcanzado el quórum mínimo de agencias y evitando que se haga busy waiting.
Además de que se puedan soportar múltiples rondas/sorteos sin necesidad de reiniciar el servidor, ya que la barrera se puede reutilizar para cada ronda/sorteo, esto es
una vez se alcanzó el quorum, se liberan los hilos y se vuelve a esperar a que se alcance el quorum para la siguiente ronda/sorteo.

## Cierre limpio de la aplicación
