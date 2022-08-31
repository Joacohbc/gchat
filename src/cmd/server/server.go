/*
Copyright © 2022 Joacohbc <joacog48@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	users        map[string]net.Conn = make(map[string]net.Conn, 0)
	colaMensajes chan string         = make(chan string)
)

// Evalua el error y lo imprime si no es un net.ErrClosed
func socketErr(err error, args ...interface{}) {
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return
		}

		log.Println(args...)
	}
}

// Abreviacion de fmt.Sprintf()
func str(format string, args ...interface{}) string {
	return fmt.Sprintf(format+"\n", args...)
}

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Inicia el servidor y lo pone escucha para los clientes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		//Parseo todas las falgs
		protocolo, err := cmd.Flags().GetString("protocol")
		cobra.CheckErr(err)

		port, err := cmd.Flags().GetString("port")
		cobra.CheckErr(err)

		cantidadMaxUsers, err := cmd.Flags().GetInt("users-max")
		cobra.CheckErr(err)

		//Escucho en el puerto indicado
		listener, err := net.Listen(protocolo, ":"+port)
		if err != nil {
			socketErr(err, "Error al inciar el servidor:", err.Error())
			os.Exit(1)
		}
		defer listener.Close()

		log.Printf("Escuchando en el puerto %s con el protocolo %s\n", port, protocolo)

		//Inicio al funcion de lectura de los mensajes
		go func() {
			//Y en una gorutine envio
			for {
				//Espero que llegue algo al canal
				v, ok := <-colaMensajes
				if !ok {
					return
				}

				//Escribio en todos a todas las conexiones el mensaje
				for name, user := range users {
					_, err := user.Write([]byte(v))
					socketErr(err, str("Error al enviar mensaje al usuario \"%s\":%v", name, err))
				}
			}
		}()

		go func() {
			//Espero a los usuarios
			for {
				//Espero una solicitud
				s, err := listener.Accept()
				socketErr(err, "Error al aceptar un cliente:", err)

				if len(users)+1 <= cantidadMaxUsers {
					//Y lo "manejo"
					go handlerUser(s)
				} else {
					_, err := s.Write([]byte("El maximo de clientes es %v, intente nuevamente mas tarde (10s se cerrara la conexion)"))
					socketErr(err, "Error al escribir mensaje de chat lleno:", err)

					time.Sleep(10 * time.Second)
					s.Close()
				}
			}
		}()

		//Creo un canal de signal
		c := make(chan os.Signal, 1)

		//Luego le "pido" que me notifique cuando se intente
		//cerrar un programa
		signal.Notify(c, os.Interrupt)

		//Si espero  que llegue la senial de cierre
		<-c
		log.Println("Apagando servidor de chat...")

		//Si llega la seña de apagado, envio un ultimo mensaje y cierro el canal
		colaMensajes <- "Servidor > El servidor se cerrar en 5s...\n"
		close(colaMensajes)

		//Espero unos segundo para dar tiempo a terminar las goroutines
		time.Sleep(5 * time.Second)

		//Y apago el servidor
		if err := listener.Close(); err != nil {
			cobra.CheckErr(fmt.Errorf("no se pudo al apagar el servidor de chat correctamente: %s", err.Error()))
		}
		log.Println("Servidor apagado con exito")
	},
}

func init() {
	//Flags
	ServerCmd.Flags().StringP("port", "p", "8081", "Indica el puerto donde escuchara el servidor")
	ServerCmd.Flags().StringP("protocol", "P", "tcp", "Indica el protocolo que usara el chat (solo tcp y udp)")
	ServerCmd.Flags().IntP("users-max", "u", 10, "Indica la cantidad de usuarios maximos de cada chat creado")
}

// Pregunta el nombre a cada usuario y luego se pone a
// leer el canal constantemente hasta que reciba un mensaje
// cuando lo recibe lo envia a la cola de mensajes
func handlerUser(user net.Conn) {

	//Cierro la conexion al final de la funcion
	defer func() {
		if user != nil {
			user.Close()
		}
	}()

	var nombre string

	for {
		//Le pido al usuario que se ponga un nombre
		_, err := user.Write([]byte("Servidor > Ingrese un nombre de usuario\n"))
		socketErr(err, "Error al leer el mensaje de identificacion del usuario:", err)

		//Leo su respuesta
		buf := make([]byte, 1024)

		n, err := user.Read(buf)
		socketErr(err, "Error al escribir el mensaje de identificacion del usuario:", err)

		//Asigno la respueta al nombre (Borro los espacios de incio y fin del nombre)
		nombre = strings.TrimSpace(string(buf[:n]))

		if nombre == "" || strings.ContainsAny(nombre, " ") {
			_, err = user.Write([]byte("Servidor > El nombre no puede estar no vacio ni puede contener espacios\n"))
			socketErr(err, "Error al leer el mensaje de identificacion del usuario:", err)

			continue //para que repita el bucle hasta encontra un nombre valido
		}

		//Si el nombre existe, osea, el nombre es invalido que haga continue
		//para volver a preguntar el nombre
		if _, existe := users[nombre]; existe {

			//Le pido al usuario que se ponga un nombre
			_, err = user.Write([]byte("Servidor > Ya existe un usuario con ese nombre en el chat\n"))
			socketErr(err, "Error al leer el mensaje de identificacion del usuario:", err)

			continue //para que repita el bucle hasta encontra un nombre valido
		}

		//Agrego el usuario a la lista una vez elegio un nombre
		users[nombre] = user
		break
	}

	//Notifico que se conecto un usuario
	log.Printf("Se conecto %s con el nombre de: %s\n", user.RemoteAddr().String(), nombre)

	//Agrego un mensaje de bienvenida a la cola
	colaMensajes <- str("Servidor > El usuario %v se ha unido al chat\n", nombre)

	//Y incio un bucle para recibir todos los mensaje de ese usuario
	for {
		buf := make([]byte, 1024)

		//Leo le mensaje del usuario
		n, err := user.Read(buf)
		if err != nil {
			socketErr(err, "Error de lectura de mensaje:", err)
			return
		}

		if string(buf[:n]) == ".exit" {
			colaMensajes <- str("El Usuario %v se fue del chat\n", nombre)
			delete(users, nombre)
			return
		}

		//Y lo pongo en la cola de mensajes
		colaMensajes <- str("%v > %v", nombre, string(buf[:n]))
	}
}
