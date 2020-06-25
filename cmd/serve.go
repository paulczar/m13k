/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"strings"

	"github.com/paulczar/m13k/pkg/webhook"
	"github.com/spf13/cobra"
)

var (
	tlsCert       string
	tlsKey        string
	port          string
	mutateCommand string
	mutateArgs    []string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		argsLenAtDash := cmd.ArgsLenAtDash()
		if argsLenAtDash > -1 {
			mutateArgs = append(mutateArgs, args[argsLenAtDash:]...)
		}
		fmt.Println("Launching webook on port", port)
		fmt.Println("Mutation by", mutateCommand, strings.Join(mutateArgs, " "))
		webhook.Serve(port, tlsCert, tlsKey, mutateCommand, mutateArgs)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serveCmd.Flags().StringVar(&tlsCert, "cert", "", "Path to TLS Certificate")
	serveCmd.Flags().StringVar(&tlsKey, "key", "", "Path to TLS Key")
	serveCmd.Flags().StringVar(&port, "port", ":8443", "port to listen on")
	serveCmd.Flags().StringVar(&mutateCommand, "command", "cat", "command to mutate resources")
	serveCmd.Flags().StringArrayVar(&mutateArgs, "args", []string{}, "args to pass to mutate command")
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	serveCmd.MarkFlagRequired("cert")
	serveCmd.MarkFlagRequired("key")

}
