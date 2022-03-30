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
package cmd

import (
	"fmt"
	"net"
	"time"

	"github.com/spf13/cobra"
)

var (
	users        map[string]net.Conn
	colaMensajes chan string
)

// rootCmd represents the base command when called without any subcommands
var serverCmd = &cobra.Command{
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
		ln, err := net.Listen(protocolo, ":"+port)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ln.Close()

		fmt.Printf("Escuchando en el puerto %s con el protocolo %s\n", port, protocolo)

		users = make(map[string]net.Conn, 0)
		colaMensajes = make(chan string)

		//Inicio al funcion de lectura de los mensajes
		go func() {
			//Y en una gorutine envio
			for {
				//Espero que llegue algo al canal
				v, ok := <-colaMensajes
				if !ok {
					fmt.Println("Cerrando chat...")
					break
				}

				//Escribio en todos lo canales el mensaje
				for _, sok := range users {
					_, err := sok.Write([]byte(v))
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}()

		//Espero a los usuarios
		for {
			//Espero una solicitud
			s, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				break
			}

			if len(users)+1 <= cantidadMaxUsers {
				//Y lo "manejo"
				go handlerUser(s)
			} else {
				_, err := s.Write([]byte("El maximo de clientes es %v, intente nuevamente mas tarde (10s se cerrara la conexion)"))
				if err != nil {
					fmt.Println(err)
					return
				}
				time.Sleep(10 * time.Second)
				s.Close()
			}
		}

		//En caso de error cierro el canal
		close(colaMensajes)

	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("port", "p", "8081", "Indica el puerto donde escuchara el servidor")
	serverCmd.Flags().StringP("protocol", "P", "tcp", "Indica el protocolo que usara el chat (solo tcp y udp)")
	serverCmd.Flags().IntP("users-max", "u", 10, "Indica la cantidad de usuarios maximos de cada chat creado")
}

//Pregunta el nombre a cada usuario y luego se pone a
//leer el canal constantemente hasta que reciba un mensaje
//cuando lo recibe lo envia a la cola de mensajes
func handlerUser(user net.Conn) {
	//Cierro la conexion al final de la funcion
	defer user.Close()

	var nombre string
	for {
		//Le pido al usuario que se ponga un nombre
		_, err := user.Write([]byte("Servidor > Ingrese un nombre para el chat:\n"))
		if err != nil {
			fmt.Println(err)
			return
		}

		//Leo su respuesta
		buf := make([]byte, 1024)
		n, err := user.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}

		//Asigno la respueta al nombre
		nombre = string(buf[:n])

		//Si el nombre no existe, osea, es un nombre valido que salga del for y continue
		if _, existe := users[nombre]; !existe {
			//Agrego el usuario a la lista una vez elegio un nombre
			users[nombre] = user
			break
		}

		//Le pido al usuario que se ponga un nombre
		_, err = user.Write([]byte("Servidor > Ya existe un usuario con ese nombre en el chat\n"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	//Notifico que se conecto un usuario
	fmt.Printf("Se conecto %s con el nombre de: %s\n", user.RemoteAddr().String(), nombre)

	//Agrego un mensaje de bienvenida a la cola
	colaMensajes <- fmt.Sprintf("Servidor > El usuario %v se ha unido al chat\n", nombre)

	//Y incio un bucle para recibir todos los mensaje de ese usuario
	for {
		buf := make([]byte, 1024)

		//Leo le mensaje del usuario
		n, err := user.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}

		if string(buf[:n]) == "exit-chat" {
			colaMensajes <- fmt.Sprintf("El Usuario %v se fue del chat\n", nombre)
			delete(users, nombre)
			break
		}

		//Y lo pongo en la cola de mensajes
		colaMensajes <- fmt.Sprintf("%v > %v", nombre, string(buf[:n]))
		fmt.Printf("El Usuario %v escribio %v\n", nombre, string(buf[:n]))
	}
}
