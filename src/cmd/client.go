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
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Inicia la consola de cliente para conectarse a un servidor",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {

		//Funcion para cerrar de manera "segura"
		stop := func(app *tview.Application, con net.Conn, err error) {
			app.Stop()
			con.Close()
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		//Parseo todas las falgs
		protocolo, err := cmd.Flags().GetString("protocol")
		cobra.CheckErr(err)

		port, err := cmd.Flags().GetString("port")
		cobra.CheckErr(err)

		ip, err := cmd.Flags().GetIP("ip")
		cobra.CheckErr(err)

		//Abro el Socket
		l, err := net.Dial(protocolo, ip.String()+":"+port)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer l.Close()

		//Creo la App
		app := tview.NewApplication()

		//Creo TextView
		textView := tview.NewTextView()
		{
			textView.SetTitle("Go Chats")
		}

		//Creo el Input
		input := tview.NewInputField()
		{
			input.SetLabel("> ")
			input.SetFieldWidth(0)
			input.SetDoneFunc(func(key tcell.Key) {

				//Cuando el a Enter
				if key == tcell.KeyEnter {
					//Que vacia el TextBox
					defer input.SetText("")

					//Si ingresa clear que vacie el TextView
					if input.GetText() == "clear" {
						textView.Clear()
						return
					}

					//Si ingresa exite que salga
					if input.GetText() == "exit" {
						app.Stop()
						os.Exit(0)
						return
					}

					//Si ingresa algo de texto
					if strings.TrimSpace(input.GetText()) != "" {

						//Que lo envie por el Socket
						_, err = l.Write([]byte(input.GetText()))
						if err != nil {
							stop(app, l, err)
						}
					}
				}
			})
		}

		//Agrego un Flex
		flex := tview.NewFlex().SetDirection(tview.FlexColumn)
		{
			flex.AddItem(textView, 0, 1, false)
			flex.AddItem(input, 0, 1, true)
		}

		//Y envio la funcion que Recibe los mensajes en segundo plano (go routine)
		go func() {
			/*
				En un bucle infinito pongo Leo todos los mensajes entrantes y
				los imprimo (y actualizo).

				Y repito otra vez esperando el siguiente mensaje
			*/
			for {
				buf := make([]byte, 1024)

				n, err := l.Read(buf)
				if err != nil {
					stop(app, l, err)
				}

				textView.SetText(textView.GetText(false) + "\n" + string(buf[:n]))
				app.ForceDraw()
			}
		}()

		if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
			stop(app, l, err)
		}

	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.Flags().StringP("port", "p", "8081", "Indica el puerto donde escuchara el servidor")
	clientCmd.Flags().StringP("protocol", "P", "tcp", "Indica el protocolo que usara el chat (solo tcp y udp)")
	clientCmd.Flags().IP("ip", []byte{byte(127), byte(0), byte(0), byte(1)}, "Indicia el IP del servidor de chat")
}
