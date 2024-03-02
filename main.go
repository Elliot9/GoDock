package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
)

type commandType interface {
	GetName() string        // Added getName method for consistency
	GetDescription() string // Added getDescription method for consistency
	Execute()
}

type concreteCommand struct {
	name        string
	description string
	executeFn   func()
}

func (c concreteCommand) GetName() string {
	return c.name
}

// GetDescription returns the command's description
func (c concreteCommand) GetDescription() string {
	return c.description
}

// Execute executes the command's functionality
func (c concreteCommand) Execute() {
	c.executeFn()
}

func makeStartCommand() commandType {
	return concreteCommand{
		name:        "s",
		description: "啟動",
		executeFn: func() {
			command := exec.Command("bash", "-c", "docker compose -f docker-compose.yml up -d")
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			err := command.Run()
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
}

func makeReStartCommand() commandType {
	return concreteCommand{
		name:        "s2",
		description: "重啟",
		executeFn: func() {
			command := exec.Command("bash", "-c", "docker compose -f docker-compose.yml up --build -d")
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			err := command.Run()
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
}

func makeCloseCommand() commandType {
	return concreteCommand{
		name:        "c",
		description: "關閉",
		executeFn: func() {
			command := exec.Command("bash", "-c", "docker compose -f docker-compose.yml down -v")
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			err := command.Run()
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
}

func makeEnterCommand() commandType {
	return concreteCommand{
		name:        "e",
		description: "進入",
		executeFn: func() {
			done := make(chan bool)
			go func() {
				cmd := exec.Command("docker", "ps", "--format", "{{.Names}}")
				output, err := cmd.Output()
				if err != nil {
					fmt.Println("Error getting container names:", err)
					done <- false
					return
				}

				containers := strings.Split(string(output), "\n")
				containers = removeEmptyStrings(containers)

				containersStr := strings.Join(containers, "、")
				reader := bufio.NewReader(os.Stdin)
				fmt.Printf("請輸入要進入的容器(%s): ", containersStr)
				container, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading user input:", err)
					done <- false
					return
				}
				container = strings.TrimSpace(container)

				if !containsString(containers, container) {
					fmt.Println("無此容器")
					done <- false
					return
				}

				// Create subcommand inside the function and use Wait
				subcommand := exec.Command("docker", "exec", "-it", container, "bash")
				subcommand.Stdin = os.Stdin
				subcommand.Stdout = os.Stdout
				subcommand.Stderr = os.Stderr

				err = subcommand.Run()
				if err != nil {
					fmt.Println("Error entering container:", err)
				}
				done <- true
			}()
			<-done
		},
	}
}

func makeQuitCommand() commandType {
	return concreteCommand{
		name:        "q",
		description: "離開",
		executeFn: func() {
			syscall.Exit(0)
		},
	}
}

var commandMaps map[string]commandType

func init() {
	commandMaps = map[string]commandType{
		"s":  makeStartCommand(),
		"s2": makeReStartCommand(),
		"c":  makeCloseCommand(),
		"e":  makeEnterCommand(),
		"q":  makeQuitCommand(),
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Chdir(dir)

	uid := os.Getuid()
	gid := os.Getgid()
	pwInfo, _ := user.LookupId(fmt.Sprintf("%d", uid))

	_ = os.Setenv("USER_ID", fmt.Sprintf("%d", uid))
	_ = os.Setenv("GID", fmt.Sprintf("%d", gid))
	_ = os.Setenv("MY_NAME", pwInfo.Username)
}

func printfMenu() {
	fmt.Println("Welcome to GoDock")
	for _, command := range commandMaps {
		fmt.Printf("%s) %s\n", command.GetName(), command.GetDescription())
	}
}

func removeEmptyStrings(s []string) []string {
	var result []string
	for _, str := range s {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

func containsString(s []string, str string) bool {
	for _, elem := range s {
		if elem == str {
			return true
		}
	}
	return false
}

func main() {
	for {
		printfMenu()
		fmt.Print("請輸入指令: ")

		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			return
		}

		if cmd, ok := commandMaps[input]; ok {
			cmd.Execute()
		} else {
			fmt.Println("無效指令")
		}
	}
}
