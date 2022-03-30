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
	"sync"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Inicia el servidor y lo pone escucha para los clientes",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		//Para iniciar el servidor
		wg := sync.WaitGroup{}

		//Parseo todas las falgs
		protocolo, err := cmd.Flags().GetString("protocol")
		cobra.CheckErr(err)

		port, err := cmd.Flags().GetString("port")
		cobra.CheckErr(err)

		cantidadUsers, err := cmd.Flags().GetInt("users")
		cobra.CheckErr(err)

		//Escucho en el puerto indicado
		ln, err := net.Listen(protocolo, ":"+port)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ln.Close()

		fmt.Printf("Escuchando en el puerto %s con el protocolo %s\n", port, protocolo)

		users := make([]net.Conn, 0, cantidadUsers)

		//Espero a los usuarios
		for len(users) < cantidadUsers {
			//Espero una solicitud
			user, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Se conecto:", user.RemoteAddr())
			users = append(users, user)
		}

		fmt.Println("Empezado la asginacion de canales...")

		colaMensajes := make(chan string)

		//Creo una Gorutine para que cada usuario
		wg.Add(len(users) + 1)

		//Leer conexiones lo que hace es (en una gorutine),
		//leer un canal constantemente hasta que reciba un mensaje
		//cuando lo recibe lo envia a la colaMensajes
		leerConexiones := func(src net.Conn, i int) {
			defer src.Close()
			defer wg.Done()
			for {
				buf := make([]byte, 1024)

				n, err := src.Read(buf)
				if err != nil {
					fmt.Println(err)
					break
				}

				if string(buf[:n]) == "exit-chat" {
					close(colaMensajes)
					fmt.Printf("El Usuario %v cerro el chat\n", i)
				}

				colaMensajes <- fmt.Sprintf("Usuario %v > %v", i, string(buf[:n]))
				fmt.Printf("El Usuario %v escribio %v\n", i, string(buf[:n]))
			}
		}

		//Pongo a leer todas las conexiones
		for i := range users {
			go leerConexiones(users[i], i)
		}

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
					break
				}
			}
		}

		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("port", "p", "8081", "Indica el puerto donde escuchara el servidor")
	serverCmd.Flags().StringP("protocol", "P", "tcp", "Indica el protocolo que usara el chat (solo tcp y udp)")
	serverCmd.Flags().IntP("users", "u", 3, "Indica la cantidad de usuarios maximos de cada chat creado")
}
